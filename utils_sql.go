package wedyta

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"strconv"
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
