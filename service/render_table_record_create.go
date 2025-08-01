package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"log"
	"slices"
	"strings"
)

func (s *Service) RenderTableRecordCreate(ctx *gin.Context) {
	modelName := ctx.Param("modelName")
	action := "create"

	permit, mConfig := s.checkAccessAndLoadModelConfig(ctx, modelName, action)
	if !permit {
		return
	}

	htmlTable, err := s.renderModelTableRecordCreate(ctx, mConfig, action)
	if err != nil {
		s.SomethingWentWrong(ctx, fmt.Sprintf("RenderModelTableRecord error: %v", err))
		return
	}

	s.RenderPage(ctx, mConfig, htmlTable)
}

func (s *Service) renderModelTableRecordCreate(ctx *gin.Context, mConfig *model.ConfigOfModel, action string) (string, error) {
	if mConfig == nil {
		log.Fatalf("Wedyta: RenderModelTableRecord(): mConfig == nil")
	}

	var htmlTable strings.Builder

	htmlTable.WriteString("<div class=\"col\">\n")

	htmlTable.WriteString(`<` + s.Config.HeadersTag + `>` + mConfig.PageTitle + `</` + s.Config.HeadersTag + `>` + "\n")
	htmlTable.WriteString(s.breadcrumbBuilder(mConfig, "", action))

	htmlTable.WriteString(s.renderAddForm(ctx, mConfig, "show_record"))

	htmlTable.WriteString("</div>\n")

	return htmlTable.String(), nil
}

func (s *Service) renderAddForm(ctx *gin.Context, mConfig *model.ConfigOfModel, successfullyCreatedDestination string) string {
	if mConfig == nil || len(mConfig.AddableFields) == 0 {
		return ""
	}

	var formBuilder strings.Builder
	formBuilder.WriteString(`
` + s.Config.JQueryScriptTag + `
<script src="/wedyta/static/js/wedyta_create.js"></script>
<link rel="stylesheet" href="/wedyta/static/css/wedyta_create.css">
` + mConfig.AdditionalScripts)

	formBuilder.WriteString(fmt.Sprintf(`<form id="addForm">
        <input type="hidden" name="modelName" value="%s">`+"\n", mConfig.ModelName))

	// adding a linking field to the parent table
	if mConfig.Parent.QueryVariableName != "" && mConfig.Parent.QueryVariableValue != "" {
		// adding input type="hidden" just if input type="text" not present
		if slices.Contains(mConfig.AddableFields, mConfig.Parent.QueryVariableName) == false {
			formBuilder.WriteString(fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`+"\n", mConfig.Parent.QueryVariableName, mConfig.Parent.QueryVariableValue))
		}
	}

	formBuilder.WriteString(`<input type="hidden" name="successfullyCreatedDestination" value="` + successfullyCreatedDestination + `">` + "\n")

	countFields := 0
	for _, field := range mConfig.AddableFields {
		fldCfg := mConfig.FieldConfig[field]

		if !fldCfg.PermitDisplayInInsertMode {
			continue
		}

		if s.Config.AccessCheckFunc(ctx, mConfig.ModelName, field, "create") != true {
			continue
		}

		value := ""
		if val, exist := ctx.GetQuery(fldCfg.Field); exist {
			value = val
		}
		fmt.Printf("%s='%s'\n", fldCfg.Field, value)

		labelTag, fieldTag := s.renderFormInputTag(&fldCfg, mConfig, nil, value)

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

	formBuilder.WriteString(`<button type="submit" class="btn btn-primary">Create</button>` + "\n</form>\n")
	return formBuilder.String()
}
