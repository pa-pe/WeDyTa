package service

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/utils"
	"github.com/pa-pe/wedyta/utils/sqlutils"
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

	fmt.Printf("payload=%v\n", payload)

	id, err := getIdFromPayload(payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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

	//fixCheckboxValue(updateData)

	if !s.validateFieldValueType(ctx, mConfig, updateData) {
		return
	}

	fmt.Printf("updateData=%v\n", updateData)

	// Retrieve original values for fields to be updated
	originalData := make(map[string]interface{})
	if err := s.DB.Table(mConfig.DbTable).Where("id = ?", int64(id)).Select(allowed).Take(&originalData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve original data"})
		return
	}

	// changing dateTimeFields format
	for field, value := range originalData {
		if dateTimeFieldConfig, dateTimeFieldExists := mConfig.DateTimeFields[field]; dateTimeFieldExists {
			originalData[field] = sqlutils.ExtractFormattedTime(value, dateTimeFieldConfig)
		}
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

	for field, val := range updateData {
		fldCfg := mConfig.FieldConfig[field]
		if fldCfg.IsPassword {
			encryptedPassword, err := s.Config.EncryptPlainPasswordFunc(ctx, mConfig.DbTable, field, updateData, val.(string))
			if err != nil {
				log.Printf("HandleTableCreateRecord: Error encrypting password for field '%s': %v", field, err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			}
			updateData[field] = encryptedPassword
		}
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
