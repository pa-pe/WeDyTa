package service

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"github.com/pa-pe/wedyta/utils"
	"github.com/pa-pe/wedyta/utils/sqlutils"
	"log"
	"os"
	"regexp"
	"strings"
)

var summernoteInitTags = "<link href=\"https://cdn.jsdelivr.net/npm/summernote@0.9.1/dist/summernote.min.css\" rel=\"stylesheet\">\n<script src=\"https://cdn.jsdelivr.net/npm/summernote@0.9.1/dist/summernote.min.js\"></script>\n<script src=\"/wedyta/static/js/wedyta_init_summernote.js\"></script>\n"

func (s *Service) loadModelConfig(ctx *gin.Context, modelName string, payload map[string]interface{}) *model.ConfigOfModel {
	configPath := s.Config.ConfigDir + "/" + modelName + ".json"

	stat, err := os.Stat(configPath)
	if err != nil {
		s.SomethingWentWrong(ctx, fmt.Sprintf("Cannot stat mConfig file for model %s: %v", modelName, err))
		return nil
	}

	cached, found := s.modelCache[modelName]
	if found && cached.ModTime.Equal(stat.ModTime()) {
		s.refreshVariableDependentParams(ctx, cached.Config, payload)
		return cached.Config
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		s.SomethingWentWrong(ctx, fmt.Sprintf("No configuration found for modelName: %s, err: %v", modelName, err))
		return nil
	}

	var mConfig model.ConfigOfModel
	if err := json.Unmarshal(data, &mConfig); err != nil {
		//return nil, fmt.Errorf("failed to parse mConfig JSON: %w", err)
		s.SomethingWentWrong(ctx, fmt.Sprintf("Failed to parse mConfig JSON of modelName: %s, err: %v", modelName, err))
		return nil
	}

	mConfig.ModelName = modelName
	s.loadModelConfigDefaults(&mConfig)
	s.fillFieldConfig(&mConfig)

	if mConfig.Parent.ModelName != "" {
		//fmt.Println(parentModelName)
		mConfig.ParentConfig = s.loadModelConfig(ctx, mConfig.Parent.ModelName, payload)
		if mConfig.ParentConfig == nil {
			s.SomethingWentWrong(ctx, "Can`t load ParentConfig: "+mConfig.Parent.ModelName)
			return nil
		}
		mConfig.HasParent = true
	}

	if s.Config.VariableResolver == nil && strings.Contains(mConfig.SqlWhere, "{{") {
		s.SomethingWentWrong(ctx, fmt.Sprintf("Trying to use variables without wedytaConfig.VariableResolver modelName=%s", modelName))
		return nil
	}

	s.modelCache[modelName] = model.CachedModelConfig{
		Config:  &mConfig,
		ModTime: stat.ModTime(),
	}

	s.refreshVariableDependentParams(ctx, &mConfig, payload)

	return &mConfig
}

func (s *Service) refreshVariableDependentParams(ctx *gin.Context, mConfig *model.ConfigOfModel, payload map[string]interface{}) {
	mConfig.SqlWhere = s.resolveVariables(ctx, mConfig.ModelName, mConfig.SqlWhereOriginal)

	if mConfig.HasParent {
		mConfig.AdditionalUrlParams = s.renderAdditionalUrlParams(ctx, mConfig, payload)
	}
}

func (s *Service) renderAdditionalUrlParams(ctx *gin.Context, mConfig *model.ConfigOfModel, payload map[string]interface{}) string {
	additionalUrlParams := "?"

	if mConfig.HasParent {
		additionalUrlParams = s.renderAdditionalUrlParams(ctx, mConfig.ParentConfig, payload)
	}

	if mConfig.Parent.QueryVariableName != "" {
		queryVariableName := mConfig.Parent.QueryVariableName
		//fmt.Println("queryVariableName=" + queryVariableName)
		if payload != nil {
			if queryVariableValue, exists := payload[queryVariableName].(string); exists {
				mConfig.Parent.QueryVariableValue = queryVariableValue
			}
		} else {
			if queryVariableValue, exist := ctx.GetQuery(queryVariableName); exist {
				//fmt.Println("queryVariableValue=" + queryVariableValue)
				mConfig.Parent.QueryVariableValue = queryVariableValue
			}
		}

		if additionalUrlParams != "?" {
			additionalUrlParams += "&"
		}

		additionalUrlParams += queryVariableName + "=" + mConfig.Parent.QueryVariableValue
	}

	//fmt.Printf("m=%s, additionalUrlParams='%s'\n", mConfig.ModelName, additionalUrlParams)
	return additionalUrlParams
}

func (s *Service) loadModelConfigDefaults(mConfig *model.ConfigOfModel) {
	if mConfig.PageTitle == "" {
		mConfig.PageTitle = mConfig.ModelName
	}
	if mConfig.DbTable == "" {
		mConfig.DbTable = utils.CamelToSnake(mConfig.ModelName)
	}

	var err error
	mConfig.DbTablePrimaryKey, err = sqlutils.GetPrimaryKeyFieldName(s.DB, mConfig.DbTable)
	if err != nil {
		log.Printf("WeDyTa: can't determine primary key for table %s: %v", mConfig.DbTable, err)
	}

	mConfig.HeaderTags = `<link rel="stylesheet" href="/wedyta/static/css/wedyta.css">` + "\n"
}

func (s *Service) fillFieldConfig(mConfig *model.ConfigOfModel) {
	if mConfig.FieldConfig == nil {
		mConfig.FieldConfig = make(map[string]model.FieldParams)
	}

	// RelatedData
	for field, relatedData := range mConfig.RelatedData {
		//relatedData = utils.CamelToSnake(relatedData)
		parts := strings.Split(relatedData, ".")
		tableName := parts[0]
		fieldName := parts[1]

		if tableName != "" && fieldName != "" {

			pkField, err := sqlutils.GetPrimaryKeyFieldName(s.DB, tableName)
			if err != nil {
				log.Printf("WeDyTa: can't determine primary key for table %s: %v", tableName, err)
			} else {
				param := mConfig.FieldConfig[field]
				param.RelatedData = &model.RelatedDataConfig{
					TableAndField:       relatedData,
					TableName:           tableName,
					FieldName:           fieldName,
					PrimaryKeyFieldName: pkField,
				}
				mConfig.FieldConfig[field] = param
			}
		}
	}

	for _, field := range mConfig.Fields {
		header := mConfig.Headers[field]
		if header == "" {
			header = mConfig.Headers[utils.InvertCaseStyle(field)]
		}
		if header == "" {
			header = field
		}

		param := mConfig.FieldConfig[field]
		param.Field = field
		param.Header = header
		param.Title = mConfig.Titles[field]
		param.DisplayMode = mConfig.DisplayMode[field]
		if param.DisplayMode == "" || param.DisplayMode == "*" || param.DisplayMode == "all" {
			param.PermitDisplayInTableMode = true
			param.PermitDisplayInRecordMode = true
			param.PermitDisplayInUpdateMode = true
			param.PermitDisplayInInsertMode = true
		} else {
			if strings.Contains(param.DisplayMode, "table") {
				param.PermitDisplayInTableMode = true
			}
			if strings.Contains(param.DisplayMode, "record") {
				param.PermitDisplayInRecordMode = true
			}
			if strings.Contains(param.DisplayMode, "update") {
				param.PermitDisplayInUpdateMode = true
			}
			if strings.Contains(param.DisplayMode, "insert") {
				param.PermitDisplayInInsertMode = true
			}
		}

		if mConfig.FieldConfig[field].RelatedData != nil {
			param.FieldEditor = "select"
		}

		mConfig.FieldConfig[field] = param
	}

	columnTypes, _ := sqlutils.GetTableColumnTypes(s.DB, mConfig.DbTable)

	// FieldEditor
	for field, editor := range mConfig.FieldEditor {
		param := mConfig.FieldConfig[field]
		param.FieldEditor = editor
		mConfig.FieldConfig[field] = param

		if editor == "summernote" {
			if !strings.Contains(mConfig.AdditionalScripts, summernoteInitTags) {
				mConfig.AdditionalScripts += summernoteInitTags
			}
			mConfig.AdditionalScripts += summernoteConfig(mConfig.ModelName, field)
		}
	}

	// AddableFields
	for _, field := range mConfig.AddableFields {
		param := mConfig.FieldConfig[field]
		param.IsAddable = true
		if param.FieldEditor == "" {
			if sqlutils.IsLongTextType(columnTypes[field]) {
				param.FieldEditor = "textarea"
			} else {
				param.FieldEditor = "input"
			}
		}
		mConfig.FieldConfig[field] = param
	}

	// EditableFields
	for _, field := range mConfig.EditableFields {
		param := mConfig.FieldConfig[field]
		param.IsEditable = true
		if param.FieldEditor == "" {
			if sqlutils.IsLongTextType(columnTypes[field]) {
				param.FieldEditor = "textarea"
			} else {
				param.FieldEditor = "input"
			}
		}
		mConfig.FieldConfig[field] = param
	}

	// Required
	for _, field := range mConfig.RequiredFields {
		param := mConfig.FieldConfig[field]
		param.IsRequired = true
		mConfig.FieldConfig[field] = param
	}

	// Classes
	for field, class := range mConfig.Classes {
		param := mConfig.FieldConfig[field]
		param.Classes = class
		mConfig.FieldConfig[field] = param
	}

	// Link presets
	for field, linkConfig := range mConfig.Links {
		if linkConfig.Preset == "self" {
			linkConfig.Template = "/wedyta/" + mConfig.ModelName + "/$" + mConfig.DbTablePrimaryKey + "$"
			mConfig.Links[field] = linkConfig
		}
	}
}

func (s *Service) resolveVariables(ctx *gin.Context, modelName string, str string) string {
	if !strings.Contains(str, "{{") {
		return str
	}

	re := regexp.MustCompile(`{{(.*?)}}`)
	matches := re.FindAllStringSubmatch(str, -1)

	var variables []string
	for _, match := range matches {
		if len(match) > 1 {
			variables = append(variables, match[1])
		}
	}

	//queryParams := ctx.Request.URL.Query()
	for _, variable := range variables {
		value := ""
		//if queryValue, exists := queryParams[variable]; exists {
		//	value = queryValue[0]
		//} else {
		if s.Config.VariableResolver != nil {
			value = s.Config.VariableResolver(ctx, modelName, variable)
		} else {

		}

		if value != "" {
			str = strings.ReplaceAll(str, "{{"+variable+"}}", fmt.Sprintf("%v", value))
		} else {
			log.Printf("Wedyta: Can`t resolve variable='%s' for modelName=%s", variable, modelName)
		}
	}
	return str
}

func summernoteConfig(modelName, field string) string {
	config := `
<script>
$(document).ready(function() {
	initSummernote("` + field + `", "` + modelName + `");
});
</script>`

	return config
}
