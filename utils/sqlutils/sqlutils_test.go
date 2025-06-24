package sqlutils

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestGetPrimaryKeyFieldName_SQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		);
	`).Error
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	pk, err := GetPrimaryKeyFieldName(db, "test_table")
	if err != nil {
		t.Fatalf("getPrimaryKeyFieldName failed: %v", err)
	}
	if pk != "id" {
		t.Errorf("expected PK to be 'id', got '%s'", pk)
	}
}
