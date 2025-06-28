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

	if s.Config.AccessCheckFunc(ctx, modelName, "", "create") != true {
		ctx.String(http.StatusForbidden, "Forbidden RenderTable: "+modelName)
		return
	}

	config := s.loadModelConfig(ctx, modelName, payload)
	if config == nil {
		return
	}

	insertData := make(map[string]interface{})
	for _, field := range config.AddableFields {
		if value, exists := payload[field]; exists {
			insertData[utils.CamelToSnake(field)] = value
		}
	}

	if localConnectionField, exists := config.Parent["localConnectionField"]; exists {
		if queryVariableValue, exists := config.Parent["queryVariableValue"]; exists {
			insertData[localConnectionField] = queryVariableValue
		}
	}

	// check RequiredFields
	for _, requiredField := range config.RequiredFields {
		if value, exists := payload[requiredField]; !exists || value == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Field '%s' is required", requiredField)})
			return
		}
	}

	// check NoZeroValueFields
	for _, noZeroField := range config.NoZeroValueFields {
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

	if err := s.DB.Table(config.DbTable).Create(insertData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Service) RenderAddForm(ctx *gin.Context, mConfig *model.ModelConfig) string {
	if mConfig == nil || len(mConfig.AddableFields) == 0 {
		return ""
	}

	var formBuilder strings.Builder
	formBuilder.WriteString(`<script src="/wedyta/static/js/wedyta_create.js"></script>
<link rel="stylesheet" href="/wedyta/static/css/wedyta_create.css">
` + mConfig.AdditionalScripts)

	formBuilder.WriteString(`
	<div class="accordion" id="addFormAccordion">
        <div class="accordion-item">
            <` + s.Config.HeadersTag + ` class="accordion-header" id="addFormHeading">
                <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#addFormCollapse" aria-expanded="false" aria-controls="addFormCollapse">
                    <i class="bi-plus-square"></i> &nbsp; Add New Record
                </button>
            </` + s.Config.HeadersTag + `>
            <div id="addFormCollapse" class="accordion-collapse collapse" aria-labelledby="addFormHeading" data-bs-parent="#addFormAccordion">
                <div class="accordion-body" style="background: rgba(128,128,128,0.1);">
`)

	formBuilder.WriteString(fmt.Sprintf(`<form id="addForm">
        <input type="hidden" name="modelName" value="%s">`+"\n", mConfig.ModelName))

	if queryVariableName, exists := mConfig.Parent["queryVariableName"]; exists {
		if queryVariableValue, exists := mConfig.Parent["queryVariableValue"]; exists {
			formBuilder.WriteString(fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`+"\n", queryVariableName, queryVariableValue))
		}
	}

	countFields := 0
	for _, field := range mConfig.AddableFields {
		if !mConfig.FieldConfig[field].PermitDisplayInInsertMode {
			continue
		}

		if s.Config.AccessCheckFunc(ctx, mConfig.ModelName, field, "create") != true {
			continue
		}

		fldCfg := mConfig.FieldConfig[field]

		label := mConfig.FieldConfig[field].Header

		requiredAttr := ""
		requiredLabel := ""
		for _, requiredField := range mConfig.RequiredFields {
			if requiredField == field {
				requiredAttr = "required"
				requiredLabel = ` <span class="required-label">(required)</span>`
				break
			}
		}

		formBuilder.WriteString(fmt.Sprintf(`<div class="mb-3">
        <label for="%s" class="form-label">%s%s</label>`, field, label, requiredLabel))

		switch fldCfg.FieldEditor {
		case "textarea":
			formBuilder.WriteString(fmt.Sprintf(`<textarea class="form-control" id="%s" name="%s" %s></textarea>`, field, field, requiredAttr))
		case "input":
			formBuilder.WriteString(fmt.Sprintf("<input class=\"form-control\" type=\"text\" id=\"%s\" name=\"%s\" value=\"\">", field, field))
		case "select":
			htmlSelect, err := s.RenderRelatedDataSelect(fldCfg.RelatedData, 0)
			if err != nil {
				formBuilder.WriteString("oops")
			} else {
				formBuilder.WriteString(htmlSelect)
			}
		case "summernote":
			formBuilder.WriteString(fmt.Sprintf("<textarea class=\"form-control\" id=\"%s\" name=\"%s\" %s></textarea>", field, field, requiredAttr))
		default:
			formBuilder.WriteString("oops, something went wrong")
		}

		formBuilder.WriteString(`</div>`)

		countFields++
	}

	// skip return add form if no addable fields by AccessCheckFunc
	if countFields == 0 {
		return ""
	}

	formBuilder.WriteString(`<button type="submit" class="btn btn-primary">Add</button>` + "\n</form>\n")
	formBuilder.WriteString("</div>\n</div>\n</div>\n")
	return formBuilder.String()
}
