package service

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/utils"
)

func (s *Service) HandleTableCreateRecord(ctx *gin.Context) {
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

	action := "create"
	permit, mConfig := s.checkAccessAndLoadModelConfig(ctx, modelName, action)
	if !permit {
		return
	}

	insertData := make(map[string]interface{})
	for _, field := range mConfig.AddableFields {
		if value, exists := payload[field]; exists {
			insertData[utils.CamelToSnake(field)] = value
		}
	}

	if mConfig.Parent.LocalConnectionField != "" && mConfig.Parent.QueryVariableValue != "" {
		insertData[mConfig.Parent.LocalConnectionField] = mConfig.Parent.QueryVariableValue
	}

	// check RequiredFields
	for _, requiredField := range mConfig.RequiredFields {
		if value, exists := payload[requiredField]; !exists || value == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("ValueField '%s' is required", requiredField)})
			return
		}
	}

	// check NoZeroValueFields
	for _, noZeroField := range mConfig.NoZeroValueFields {
		if value, exists := payload[noZeroField]; exists {
			if number, ok := value.(float64); ok && number == 0 {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("ValueField '%s' cannot be zero", noZeroField)})
				return
			}
		}
	}

	if len(insertData) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No data to insert"})
		return
	}

	//fixCheckboxValue(insertData)

	if !s.validateFieldValueType(ctx, mConfig, insertData) {
		return
	}

	if s.Config.BeforeCreate != nil {
		permitCreate, msg := s.Config.BeforeCreate(ctx, s.DB, mConfig.DbTable, insertData)
		if !permitCreate {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
	}

	for field, val := range insertData {
		fldCfg := mConfig.FieldConfig[field]
		if fldCfg.IsPassword {
			encryptedPassword, err := s.Config.EncryptPlainPasswordFunc(ctx, mConfig.DbTable, field, insertData, val.(string))
			if err != nil {
				log.Printf("HandleTableCreateRecord: Error encrypting password for field '%s': %v", field, err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			}
			insertData[field] = encryptedPassword
		}
	}

	var insertedID int64
	if err := s.DB.Table(mConfig.DbTable).Create(insertData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data"})
		return
	} else {
		s.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&insertedID)
	}

	successfullyCreatedDestination := ""

	if value, exists := payload["successfullyCreatedDestination"]; exists {
		successfullyCreatedDestination = value.(string)
		if successfullyCreatedDestination == "show_record" {
			successfullyCreatedDestination = "/wedyta/" + mConfig.ModelName + "/" + strconv.FormatInt(insertedID, 10)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "successfullyCreatedDestination": successfullyCreatedDestination})
}
