package wedyta

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (c *Impl) HandleTableCreateRecord(ctx *gin.Context) {
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

	if c.Config.AccessCheckFunc(ctx, modelName, "", "create") != true {
		ctx.String(http.StatusForbidden, "Forbidden RenderTable: "+modelName)
		return
	}

	config := c.loadModelConfig(ctx, modelName, payload)
	if config == nil {
		return
	}

	insertData := make(map[string]interface{})
	for _, field := range config.AddableFields {
		if value, exists := payload[field]; exists {
			insertData[CamelToSnake(field)] = value
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

	if err := c.DB.Table(config.DbTable).Create(insertData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (c *Impl) RenderAddForm(ctx *gin.Context, config *modelConfig, modelName string) string {
	if config == nil || len(config.AddableFields) == 0 {
		return ""
	}

	var formBuilder strings.Builder
	formBuilder.WriteString(`<script src="/wedyta/static/js/wedyta_create.js"></script>
<link rel="stylesheet" href="/wedyta/static/css/wedyta_create.css">
	<div class="accordion" id="addFormAccordion">
        <div class="accordion-item">
            <` + c.Config.HeadersTag + ` class="accordion-header" id="addFormHeading">
                <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#addFormCollapse" aria-expanded="false" aria-controls="addFormCollapse">
                    <i class="bi-plus-square"></i> &nbsp; Add New Record
                </button>
            </` + c.Config.HeadersTag + `>
            <div id="addFormCollapse" class="accordion-collapse collapse" aria-labelledby="addFormHeading" data-bs-parent="#addFormAccordion">
                <div class="accordion-body" style="background: rgba(128,128,128,0.1);">
`)

	formBuilder.WriteString(fmt.Sprintf(`<form id="addForm">
        <input type="hidden" name="modelName" value="%s">`+"\n", modelName))

	if queryVariableName, exists := config.Parent["queryVariableName"]; exists {
		if queryVariableValue, exists := config.Parent["queryVariableValue"]; exists {
			formBuilder.WriteString(fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`+"\n", queryVariableName, queryVariableValue))
		}
	}

	countFields := 0
	for _, field := range config.AddableFields {
		if c.Config.AccessCheckFunc(ctx, modelName, field, "create") != true {
			continue
		}

		label := config.Headers[field]
		if label == "" {
			label = field
		}

		requiredAttr := ""
		requiredLabel := ""
		for _, requiredField := range config.RequiredFields {
			if requiredField == field {
				requiredAttr = "required"
				requiredLabel = ` <span class="required-label">(required)</span>`
				break
			}
		}

		formBuilder.WriteString(fmt.Sprintf(`<div class="mb-3">
        <label for="%s" class="form-label">%s%s</label>
        <textarea type="text" class="form-control" id="%s" name="%s" %s></textarea>
    </div>`, field, label, requiredLabel, field, field, requiredAttr))
		//		<input type="text" class="form-control" id="%s" name="%s" %s>

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
