package sqlutils

import (
	"database/sql"
	"fmt"
	"github.com/pa-pe/wedyta/model"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
	"time"
)

type ColumnSchema struct {
	Name         string
	Type         string
	IsPrimaryKey bool
	IsNullable   bool
	DefaultValue any
}

var tableSchemaCache = make(map[string][]ColumnSchema)

func getTableSchema(db *gorm.DB, tableName string) ([]ColumnSchema, error) {
	if schema, ok := tableSchemaCache[tableName]; ok {
		return schema, nil
	}

	var schema []ColumnSchema
	dialector := db.Dialector.Name()

	switch dialector {
	case "mysql":
		query := `
			SELECT COLUMN_NAME, DATA_TYPE, COLUMN_KEY, IS_NULLABLE, COLUMN_DEFAULT
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			  AND TABLE_NAME = ?
		`
		type mysqlCol struct {
			ColumnName    string         `gorm:"column:COLUMN_NAME"`
			DataType      string         `gorm:"column:DATA_TYPE"`
			ColumnKey     string         `gorm:"column:COLUMN_KEY"`
			IsNullable    string         `gorm:"column:IS_NULLABLE"`
			ColumnDefault sql.NullString `gorm:"column:COLUMN_DEFAULT"`
		}
		var results []mysqlCol
		if err := db.Raw(query, tableName).Scan(&results).Error; err != nil {
			return nil, err
		}
		//fmt.Printf("%+v\n", results) // dbg
		for _, col := range results {
			schema = append(schema, ColumnSchema{
				Name:         col.ColumnName,
				Type:         strings.ToLower(col.DataType),
				IsPrimaryKey: col.ColumnKey == "PRI",
				IsNullable:   col.IsNullable == "YES",
				DefaultValue: safeString(col.ColumnDefault),
			})
		}

	case "postgres":
		query := `
			SELECT
				a.attname AS column_name,
				format_type(a.atttypid, a.atttypmod) AS data_type,
				(i.indisprimary IS TRUE) AS is_primary,
				(a.attnotnull IS FALSE) AS is_nullable,
				pg_get_expr(ad.adbin, ad.adrelid) AS column_default
			FROM pg_attribute a
			LEFT JOIN pg_index i ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey) AND i.indisprimary
			LEFT JOIN pg_attrdef ad ON a.attrelid = ad.adrelid AND a.attnum = ad.adnum
			JOIN pg_class c ON a.attrelid = c.oid
			JOIN pg_namespace n ON c.relnamespace = n.oid
			WHERE c.relname = $1 AND a.attnum > 0 AND NOT a.attisdropped
		`
		type pgCol struct {
			ColumnName    string
			DataType      string
			IsPrimaryKey  bool
			IsNullable    bool
			ColumnDefault sql.NullString
		}
		var results []pgCol
		if err := db.Raw(query, tableName).Scan(&results).Error; err != nil {
			return nil, err
		}
		for _, col := range results {
			schema = append(schema, ColumnSchema{
				Name:         col.ColumnName,
				Type:         strings.ToLower(col.DataType),
				IsPrimaryKey: col.IsPrimaryKey,
				IsNullable:   col.IsNullable,
				DefaultValue: safeString(col.ColumnDefault),
			})
		}

	case "sqlite", "sqlite3":
		type pragmaInfo struct {
			Name         string         `gorm:"column:name"`
			Type         string         `gorm:"column:type"`
			NotNull      int            `gorm:"column:notnull"`
			DefaultValue sql.NullString `gorm:"column:dflt_value"`
			PK           int            `gorm:"column:pk"`
		}
		var results []pragmaInfo
		query := fmt.Sprintf("PRAGMA table_info(`%s`);", tableName)
		if err := db.Raw(query).Scan(&results).Error; err != nil {
			return nil, err
		}
		for _, col := range results {
			//fmt.Printf("SQLite col: %+v\n", col) // dbg
			schema = append(schema, ColumnSchema{
				Name:         col.Name,
				Type:         strings.ToLower(col.Type),
				IsPrimaryKey: col.PK > 0,
				IsNullable:   col.NotNull == 0,
				DefaultValue: col.DefaultValue,
			})
		}

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", dialector)
	}

	tableSchemaCache[tableName] = schema
	return schema, nil
}

func getPrimaryKeyFieldNameFromSchema(schema []ColumnSchema) (string, error) {
	for _, col := range schema {
		if col.IsPrimaryKey {
			return col.Name, nil
		}
	}
	return "", fmt.Errorf("primary key not found in schema")
}

func GetPrimaryKeyFieldName(db *gorm.DB, tableName string) (string, error) {
	schema, err := getTableSchema(db, tableName)
	if err != nil {
		return "", err
	}
	return getPrimaryKeyFieldNameFromSchema(schema)
}

func GetTableColumnTypes(db *gorm.DB, tableName string) (map[string]string, error) {
	schema, err := getTableSchema(db, tableName)
	if err != nil {
		return nil, err
	}

	types := make(map[string]string, len(schema))
	for _, col := range schema {
		types[col.Name] = col.Type
	}
	return types, nil
}

func safeString(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(val)
	case string:
		return val
	case *string:
		if val != nil {
			return *val
		}
		return ""
	default:
		return fmt.Sprint(val)
	}
}

func IsLongTextType(fieldType string) bool {
	fieldType = strings.ToLower(fieldType)
	return strings.Contains(fieldType, "text") || fieldType == "json" || strings.HasPrefix(fieldType, "varchar(") && ExtractFieldTypeLength(fieldType) > 255
}

func ExtractFieldTypeLength(fieldType string) int {
	// Example: varchar(500)
	start := strings.Index(fieldType, "(")
	end := strings.Index(fieldType, ")")
	if start == -1 || end == -1 || start > end {
		return 0
	}
	numStr := fieldType[start+1 : end]
	n, _ := strconv.Atoi(numStr)
	return n
}

//var primaryKeyCache = make(map[string]string)
//
//func (c *Impl) getPrimaryKeyFieldName(tableName string) (string, error) {
//	if pk, ok := primaryKeyCache[tableName]; ok {
//		return pk, nil
//	}
//
//	dialector := c.DB.Dialector.Name()
//	var columnName string
//	var err error
//
//	switch dialector {
//	case "mysql":
//		query := `
//			SELECT COLUMN_NAME
//			FROM INFORMATION_SCHEMA.COLUMNS
//			WHERE TABLE_SCHEMA = DATABASE()
//				AND TABLE_NAME = ?
//				AND COLUMN_KEY = 'PRI'
//			LIMIT 1
//		`
//		err = c.DB.Raw(query, tableName).Scan(&columnName).Error
//
//	case "sqlite", "sqlite3":
//		type tableInfo struct {
//			Cid          int
//			Name         string
//			Type         string
//			NotNull      int
//			DefaultValue any `gorm:"column:dflt_value"`
//			PK           int
//		}
//		var columns []tableInfo
//		err = c.DB.Raw(fmt.Sprintf("PRAGMA table_info(`%s`);", tableName)).Scan(&columns).Error
//		if err == nil {
//			for _, col := range columns {
//				if col.PK == 1 {
//					columnName = col.Name
//					break
//				}
//			}
//			if columnName == "" {
//				err = fmt.Errorf("no primary key found in table %s", tableName)
//			}
//		}
//	case "postgres":
//		query := `
//			SELECT a.attname
//			FROM pg_index i
//			JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
//			JOIN pg_class c ON c.oid = i.indrelid
//			JOIN pg_namespace n ON n.oid = c.relnamespace
//			WHERE i.indisprimary AND c.relname = $1
//			LIMIT 1
//		`
//		err = c.DB.Raw(query, tableName).Scan(&columnName).Error
//	default:
//		err = fmt.Errorf("unsupported database driver: %s", dialector)
//	}
//
//	if err != nil {
//		return "", err
//	}
//
//	primaryKeyCache[tableName] = columnName
//
//	return columnName, nil
//}
//
//type ColumnInfo struct {
//	Name string
//	Type string
//}
//
//func (c *Impl) getTableColumnTypes(tableName string) (map[string]string, error) {
//	columnTypes := make(map[string]string)
//
//	var results []ColumnInfo
//
//	switch c.DB.Dialector.Name() {
//	case "mysql":
//		query := `
//			SELECT COLUMN_NAME as name, DATA_TYPE as type
//			FROM INFORMATION_SCHEMA.COLUMNS
//			WHERE TABLE_SCHEMA = DATABASE()
//			  AND TABLE_NAME = ?
//		`
//		if err := c.DB.Raw(query, tableName).Scan(&results).Error; err != nil {
//			return nil, err
//		}
//	case "postgres":
//		query := `
//			SELECT column_name as name, data_type as type
//			FROM information_schema.columns
//			WHERE table_name = ?
//		`
//		if err := c.DB.Raw(query, tableName).Scan(&results).Error; err != nil {
//			return nil, err
//		}
//	case "sqlite", "sqlite3":
//		type pragmaInfo struct {
//			Name string
//			Type string
//		}
//		var pragmaResults []pragmaInfo
//		if err := c.DB.Raw(fmt.Sprintf("PRAGMA table_info(`%s`);", tableName)).Scan(&pragmaResults).Error; err != nil {
//			return nil, err
//		}
//		for _, col := range pragmaResults {
//			columnTypes[col.Name] = strings.ToLower(col.Type)
//		}
//		return columnTypes, nil
//	default:
//		return nil, fmt.Errorf("unsupported driver: %s", c.DB.Dialector.Name())
//	}
//
//	for _, col := range results {
//		columnTypes[col.Name] = strings.ToLower(col.Type)
//	}
//	return columnTypes, nil
//}

func GetTotalRecords(db *gorm.DB, config *model.ModelConfig) (int64, error) {
	var totalRecords int64
	if err := db.Table(config.DbTable).Where(config.SqlWhere).Count(&totalRecords).Error; err != nil {
		return 0, err
	}

	return totalRecords, nil
}

func IsNumericColumnType(sqlType string) bool {
	sqlType = strings.ToLower(sqlType)

	switch {
	case strings.HasPrefix(sqlType, "int"), // int, int4, int8, int(11)
		strings.HasPrefix(sqlType, "bigint"),
		strings.HasPrefix(sqlType, "smallint"),
		strings.HasPrefix(sqlType, "tinyint"),
		strings.HasPrefix(sqlType, "mediumint"),
		strings.HasPrefix(sqlType, "serial"),

		strings.HasPrefix(sqlType, "float"),
		strings.HasPrefix(sqlType, "double"),
		strings.HasPrefix(sqlType, "real"),
		strings.HasPrefix(sqlType, "numeric"),
		strings.HasPrefix(sqlType, "decimal"):
		return true
	default:
		return false
	}
}

func SanitizeNumericField(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil, false
		}
		// try to reduce it to a number
		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return f, true
		}
		return nil, false

	case float64, float32, int, int64, int32, uint, uint64:
		// already a number
		return v, true

	default:
		return nil, false
	}
}

func ExtractInt64(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int8:
		return int64(v)
	case uint8:
		return int64(v)
	case int32:
		return int64(v)
	case uint64:
		return int64(v)
	case uint32:
		return int64(v)
	case int:
		return int64(v)
	case uint:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case []byte:
		if num, err := strconv.ParseInt(string(v), 10, 64); err == nil {
			return num
		}
	case string:
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			return num
		}
	default:
		log.Printf("WeDyTa TODO: unhandled type %T with value %v", value, value)
	}
	return 0
}

func ExtractFormattedTime(value any, outputFormat string) string {
	switch v := value.(type) {
	case time.Time:
		return v.Format(outputFormat)
	case string:
		for _, layout := range []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
			"02.01.2006 15:04:05",
			"02.01.2006",
		} {
			if t, err := time.Parse(layout, v); err == nil {
				return t.Format(outputFormat)
			}
		}
		return v // fallback
	case []byte:
		return ExtractFormattedTime(string(v), outputFormat)
	default:
		return fmt.Sprintf("%v", value)
	}
}
