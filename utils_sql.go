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

	var columnName string
	query := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND COLUMN_KEY = 'PRI'
		LIMIT 1
	`
	err := db.Raw(query, tableName).Scan(&columnName).Error
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
