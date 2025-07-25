package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"github.com/pa-pe/wedyta/utils"
	"net/http"
	"strings"
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

	if localConnectionField, exists := mConfig.Parent["localConnectionField"]; exists {
		if queryVariableValue, exists := mConfig.Parent["queryVariableValue"]; exists {
			insertData[localConnectionField] = queryVariableValue
		}
	}

	// check RequiredFields
	for _, requiredField := range mConfig.RequiredFields {
		if value, exists := payload[requiredField]; !exists || value == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Field '%s' is required", requiredField)})
			return
		}
	}

	// check NoZeroValueFields
	for _, noZeroField := range mConfig.NoZeroValueFields {
		if value, exists := payload[noZeroField]; exists {
			if number, ok := value.(float64); ok && number == 0 {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Field '%s' cannot be zero", noZeroField)})
				return
			}
		}
	}

	if len(insertData) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No data to insert"})
		return
	}

	if err := s.DB.Table(mConfig.DbTable).Create(insertData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Service) RenderAddForm(ctx *gin.Context, mConfig *model.ConfigOfModel) string {
	if mConfig == nil || len(mConfig.AddableFields) == 0 {
		return ""
	}

	var formBuilder strings.Builder
	formBuilder.WriteString(`<script src="/wedyta/static/js/wedyta_create.js"></script>
<link rel="stylesheet" href="/wedyta/static/css/wedyta_create.css">
` + mConfig.AdditionalScripts)

	formBuilder.WriteString(fmt.Sprintf(`<form id="addForm">
        <input type="hidden" name="modelName" value="%s">`+"\n", mConfig.ModelName))

	if queryVariableName, exists := mConfig.Parent["queryVariableName"]; exists {
		if queryVariableValue, exists := mConfig.Parent["queryVariableValue"]; exists {
			formBuilder.WriteString(fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`+"\n", queryVariableName, queryVariableValue))
		}
	}

	countFields := 0
	for _, field := range mConfig.AddableFields {
		fldCfg := mConfig.FieldConfig[field]

		if !fldCfg.PermitDisplayInInsertMode {
			continue
		}

		if s.Config.AccessCheckFunc(ctx, mConfig.ModelName, field, "create") != true {
			continue
		}

		labelTag, fieldTag := s.renderFormInputTag(&fldCfg, nil, "")

		formBuilder.WriteString("<div class=\"mb-3\">\n")
		formBuilder.WriteString(labelTag + "\n")
		formBuilder.WriteString(fieldTag + "\n")
		formBuilder.WriteString("</div>\n")

		countFields++
	}

	// skip return add form if no addable fields by AccessCheckFunc or no PermitDisplayInInsertMode
	if countFields == 0 {
		return ""
	}

	formBuilder.WriteString(`<button type="submit" class="btn btn-primary">Add</button>` + "\n</form>\n")
	return formBuilder.String()
}
