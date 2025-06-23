package wedyta

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
	"time"
)

var primaryKeyCache = make(map[string]string)

func getPrimaryKeyFieldName(db *gorm.DB, tableName string) (string, error) {
	if pk, ok := primaryKeyCache[tableName]; ok {
		return pk, nil
	}

	dialector := db.Dialector.Name()
	var columnName string
	var err error

	switch dialector {
	case "mysql":
		query := `
			SELECT COLUMN_NAME
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
				AND TABLE_NAME = ?
				AND COLUMN_KEY = 'PRI'
			LIMIT 1
		`
		err = db.Raw(query, tableName).Scan(&columnName).Error

	case "sqlite", "sqlite3":
		type tableInfo struct {
			Cid          int
			Name         string
			Type         string
			NotNull      int
			DefaultValue any `gorm:"column:dflt_value"`
			PK           int
		}
		var columns []tableInfo
		err = db.Raw(fmt.Sprintf("PRAGMA table_info(`%s`);", tableName)).Scan(&columns).Error
		if err == nil {
			for _, col := range columns {
				if col.PK == 1 {
					columnName = col.Name
					break
				}
			}
			if columnName == "" {
				err = fmt.Errorf("no primary key found in table %s", tableName)
			}
		}
	case "postgres":
		query := `
			SELECT a.attname
			FROM pg_index i
			JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
			JOIN pg_class c ON c.oid = i.indrelid
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE i.indisprimary AND c.relname = $1
			LIMIT 1
		`
		err = db.Raw(query, tableName).Scan(&columnName).Error
	default:
		err = fmt.Errorf("unsupported database driver: %s", dialector)
	}

	if err != nil {
		return "", err
	}

	primaryKeyCache[tableName] = columnName

	return columnName, nil
}

type ColumnInfo struct {
	Name string
	Type string
}

func getTableColumnTypes(db *gorm.DB, tableName string) (map[string]string, error) {
	columnTypes := make(map[string]string)

	var results []ColumnInfo

	switch db.Dialector.Name() {
	case "mysql":
		query := `
			SELECT COLUMN_NAME as name, DATA_TYPE as type
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			  AND TABLE_NAME = ?
		`
		if err := db.Raw(query, tableName).Scan(&results).Error; err != nil {
			return nil, err
		}
	case "postgres":
		query := `
			SELECT column_name as name, data_type as type
			FROM information_schema.columns
			WHERE table_name = ?
		`
		if err := db.Raw(query, tableName).Scan(&results).Error; err != nil {
			return nil, err
		}
	case "sqlite", "sqlite3":
		type pragmaInfo struct {
			Name string
			Type string
		}
		var pragmaResults []pragmaInfo
		if err := db.Raw(fmt.Sprintf("PRAGMA table_info(`%s`);", tableName)).Scan(&pragmaResults).Error; err != nil {
			return nil, err
		}
		for _, col := range pragmaResults {
			columnTypes[col.Name] = strings.ToLower(col.Type)
		}
		return columnTypes, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", db.Dialector.Name())
	}

	for _, col := range results {
		columnTypes[col.Name] = strings.ToLower(col.Type)
	}
	return columnTypes, nil
}

func (c *Impl) getTotalRecords(db *gorm.DB, config *modelConfig) (int64, error) {
	var totalRecords int64
	if err := db.Debug().Table(config.DbTable).Where(config.SqlWhere).Count(&totalRecords).Error; err != nil {
		return 0, err
	}

	return totalRecords, nil
}

func isNumericColumnType(sqlType string) bool {
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

func sanitizeNumericField(value interface{}) (interface{}, bool) {
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

func extractInt64(value any) int64 {
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

func extractFormattedTime(value any, outputFormat string) string {
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
		return extractFormattedTime(string(v), outputFormat)
	default:
		return fmt.Sprintf("%v", value)
	}
}
