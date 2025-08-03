package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"github.com/pa-pe/wedyta/utils/sqlutils"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func extractIsActive(record map[string]interface{}) bool {
	if isActive, exists := record["is_active"]; exists {
		return fmt.Sprint(isActive) == "1"
	}

	if isDisabled, exists := record["is_disabled"]; exists {
		return fmt.Sprint(isDisabled) == "0"
	}

	return true
}

func getIdFromPayload(payload map[string]interface{}) (float64, error) {
	id, ok := payload["id"].(float64)
	if !ok {
		idStr, ok := payload["id"].(string)
		if !ok {
			return 0, errors.New("ID is required")
		}

		idFromStr, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return 0, errors.New("ID recognize")
		}
		id = float64(idFromStr)
	}

	return id, nil
}

func fixCheckboxValue(data map[string]interface{}) {
	for field, val := range data {
		if field == "is_active" {
			strVal := fmt.Sprint(val)
			switch strings.ToLower(strVal) {
			case "on", "1", "true":
				val = 1
			default:
				val = 0
			}
			data[field] = val
		}
	}
}

func (s *Service) validateFieldValueType(ctx *gin.Context, mConfig *model.ConfigOfModel, data map[string]interface{}) bool {
	fieldTypes, err := sqlutils.GetTableColumnTypes(s.DB, mConfig.DbTable)
	if err != nil {
		log.Printf("Wedyta: getTableColumnTypes() error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return false
	}

	for field, val := range data {
		colType, ok := fieldTypes[field]
		if !ok {
			log.Printf("Wedyta: ValueField '%s' not found in TableColumnTypes", field)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return false
		}

		if sqlutils.IsNumericColumnType(colType) {
			//log.Printf("dbg isNumeric: %s", field)

			cleaned, ok := sqlutils.SanitizeNumericField(val)
			if !ok {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("ValueField '%s' expects a numeric value", mConfig.FieldConfig[field].Header)})
				return false
			}

			original := fmt.Sprint(val)
			cleanedStr := fmt.Sprint(cleaned)

			if original != cleanedStr {
				if strings.TrimSpace(original) == cleanedStr {
					data[field] = strings.TrimSpace(original)
				} else {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("ValueField '%s' has invalid formatting (spaces or extra characters)", field)})
					return false
				}
			}
		}
	}

	return true
}
