package sqlutils

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"reflect"
	"testing"
)

var testDBSQLite *gorm.DB

// TestMain_SQLite инициализирует SQLite in-memory базу данных для всех тестов в этом пакете.
func TestMain(m *testing.M) {
	var err error
	testDBSQLite, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to open sqlite test database: " + err.Error())
	}

	os.Exit(m.Run())
}

func TestGetPrimaryKeyFieldName_SQLite_TableNotExists(t *testing.T) {
	_, err := GetPrimaryKeyFieldName(testDBSQLite, "non_existing_table")
	if err == nil {
		t.Errorf("expected error for non-existing table, got nil")
	}
}

func TestGetPrimaryKeyFieldName_SQLite_NoPK(t *testing.T) {
	err := testDBSQLite.Exec(`
		CREATE TABLE test_no_pk (
			name TEXT
		);
	`).Error
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	_, err = GetPrimaryKeyFieldName(testDBSQLite, "test_no_pk")
	if err == nil {
		t.Errorf("expected error when no primary key is present, got nil")
	}
}

func TestGetPrimaryKeyFieldName_SQLite(t *testing.T) {
	err := testDBSQLite.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		);
	`).Error
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	pk, err := GetPrimaryKeyFieldName(testDBSQLite, "test_table")
	if err != nil {
		t.Fatalf("getPrimaryKeyFieldName failed: %v", err)
	}
	if pk != "id" {
		t.Errorf("expected PK to be 'id', got '%s'", pk)
	}
}

func TestGetTableColumnTypes_SQLite(t *testing.T) {
	err := testDBSQLite.Exec(`
		CREATE TABLE test_types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			age INTEGER,
			active BOOLEAN
		);
	`).Error
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	colTypes, err := GetTableColumnTypes(testDBSQLite, "test_types")
	if err != nil {
		t.Fatalf("GetTableColumnTypes failed: %v", err)
	}

	expected := map[string]string{
		"id":     "integer",
		"name":   "text",
		"age":    "integer",
		"active": "boolean",
	}

	for col, expectedType := range expected {
		gotType, ok := colTypes[col]
		if !ok {
			t.Errorf("column %s not found in results", col)
		} else if gotType != expectedType {
			t.Errorf("expected type of %s to be %s, got %s", col, expectedType, gotType)
		}
	}
}

func TestParseRawSql(t *testing.T) {
	tests := []struct {
		raw     string
		success bool
		expect  ParsedSql
	}{
		{
			raw:     "SELECT id, name FROM users WHERE active = 1 ORDER BY name;",
			success: true,
			expect: ParsedSql{
				RawSql:  "SELECT id, name FROM users WHERE active = 1 ORDER BY name;",
				Fields:  []string{"id", "name"},
				Table:   "users",
				Where:   "active = 1",
				OrderBy: "name",
			},
		},
		{
			raw:     "SELECT user_id, username FROM web_users ORDER BY username",
			success: true,
			expect: ParsedSql{
				RawSql:  "SELECT user_id, username FROM web_users ORDER BY username",
				Fields:  []string{"user_id", "username"},
				Table:   "web_users",
				OrderBy: "username",
			},
		},
		{
			raw:     "SELECT a, b FROM table;",
			success: true,
			expect: ParsedSql{
				RawSql: "SELECT a, b FROM table;",
				Fields: []string{"a", "b"},
				Table:  "table",
			},
		},
		{
			raw:     "SELECT onlyone FROM test;",
			success: false,
		},
		{
			raw:     "invalid syntax",
			success: false,
		},
	}

	for i, tt := range tests {
		got, ok := ParseRawSql(tt.raw)
		if ok != tt.success {
			t.Errorf("test %d: expected success %v, got %v", i, tt.success, ok)
			continue
		}
		if ok {
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("test %d: expected %+v, got %+v", i, tt.expect, got)
			}
		}
	}
}
