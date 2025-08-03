package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/utils"
	"github.com/pa-pe/wedyta/utils/sqlutils"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Update updates the fields of a specified model based on the allowedFields
func (s *Service) Update(ctx *gin.Context) {
	var payload map[string]interface{}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	modelName, ok := payload["modelName"].(string)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model name is required"})
		return
	}

	if s.Config.AccessCheckFunc(ctx, modelName, "", "update") != true {
		//ctx.String(http.StatusForbidden, "Forbidden RenderTable: "+modelName)
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Access denied", "modelName": "$modelName"})
		return
	}

	mConfig := s.loadModelConfig(ctx, modelName, payload)
	if mConfig == nil {
		return
	}

	//var payload map[string]interface{}
	//if err := ctx.ShouldBindJSON(&payload); err != nil {
	//	ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
	//	return
	//}
	//
	//modelName, ok := payload["model"].(string)
	//if !ok {
	//	ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model name is required"})
	//	return
	//}

	fmt.Println(payload)

	id, ok := payload["id"].(float64)
	if !ok {
		idStr, ok := payload["id"].(string)
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		idFromStr, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID recognize"})
			return
		}
		id = float64(idFromStr)
	}

	var allowed []string

	//Prepare the map for updating
	updateData := make(map[string]interface{})
	//	if class, ok := mConfig.EditableFields[field]; ok {

	//for field := range mConfig.EditableFields {
	for _, field := range mConfig.EditableFields {
		fieldSnaked := utils.CamelToSnake(field)
		if value, exists := payload[fieldSnaked]; exists {
			updateData[fieldSnaked] = value
			allowed = append(allowed, fieldSnaked)
		} else if value, exists := payload[field]; exists {
			updateData[fieldSnaked] = value
			allowed = append(allowed, fieldSnaked)
		}
	}

	if len(updateData) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	// clean numeric fields
	fieldTypes, err := sqlutils.GetTableColumnTypes(s.DB, mConfig.DbTable)
	if err != nil {
		log.Printf("Wedyta: getTableColumnTypes() error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	for field, val := range updateData {
		colType, ok := fieldTypes[field]
		if !ok {
			log.Printf("Wedyta: ValueField '%s' not found in TableColumnTypes", field)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		if field == "is_active" {
			strVal := fmt.Sprint(val)
			switch strings.ToLower(strVal) {
			case "on", "1", "true":
				val = 1
			default:
				val = 0
			}
			updateData[field] = val
		}

		if sqlutils.IsNumericColumnType(colType) {
			//log.Printf("dbg isNumeric: %s", field)

			cleaned, ok := sqlutils.SanitizeNumericField(val)
			if !ok {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("ValueField '%s' expects a numeric value", mConfig.FieldConfig[field].Header)})
				return
			}

			original := fmt.Sprint(val)
			cleanedStr := fmt.Sprint(cleaned)

			if original != cleanedStr {
				if strings.TrimSpace(original) == cleanedStr {
					updateData[field] = strings.TrimSpace(original)
				} else {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("ValueField '%s' has invalid formatting (spaces or extra characters)", field)})
					return
				}
			}
		}
	}

	// Retrieve original values for fields to be updated
	originalData := make(map[string]interface{})
	if err := s.DB.Table(mConfig.DbTable).Where("id = ?", int64(id)).Select(allowed).Take(&originalData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve original data"})
		return
	}

	// filtering the same data
	for key, newVal := range updateData {
		if oldVal, exists := originalData[key]; exists {
			// Приводим к строке для сравнения — учитывает типы вроде []uint8 vs string
			if fmt.Sprint(oldVal) == fmt.Sprint(newVal) {
				delete(updateData, key)
				//log.Printf("field have the same data: %s", key)
			}
		}
	}

	if len(updateData) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No new data for update"})
		return
	}

	if err := s.DB.Table(mConfig.DbTable).Where(fmt.Sprint("id = ", id)).Updates(updateData).Error; err != nil {
		log.Printf("Wedyta: Failed to update model, error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update model"})
		return
	}

	if s.Config.AfterUpdate != nil {
		for field, newValue := range updateData {
			originalValue, exists := originalData[field]
			if exists && originalValue != newValue {
				go s.Config.AfterUpdate(ctx, s.DB, mConfig.DbTable, int64(id), field, fmt.Sprintf("%v", originalValue), fmt.Sprintf("%v", newValue))
			}
		}
	}

	// client side js tests:
	//time.Sleep(1000 * time.Millisecond)
	//ctx.JSON(http.StatusBadRequest, gin.H{"error": "test fail"})
	//return

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Model updated successfully"})
}
