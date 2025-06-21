package wedyta

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"regexp"
	"strings"
)

func (c *Impl) loadModelConfig(context *gin.Context, modelName string, payload map[string]interface{}) *modelConfig {
	configPath := c.Config.ConfigDir + "/" + modelName + ".json"

	data, err := os.ReadFile(configPath)
	if err != nil {
		c.somethingWentWrong(context, fmt.Sprintf("No configuration found for modelName: %s, err: %v", modelName, err))
		return nil
	}

	var config modelConfig
	if err := json.Unmarshal(data, &config); err != nil {
		//return nil, fmt.Errorf("failed to parse config JSON: %w", err)
		c.somethingWentWrong(context, fmt.Sprintf("Failed to parse config JSON of modelName: %s, err: %v", modelName, err))
		return nil
	}

	config.ModelName = modelName
	c.loadModelConfigDefaults(&config)
	c.fillFieldConfig(&config)

	if parentModelName, parentExists := config.Parent["modelName"]; parentExists {
		//fmt.Println(parentModelName)
		config.ParentConfig = c.loadModelConfig(context, parentModelName, payload)
		if config.ParentConfig == nil {
			c.somethingWentWrong(context, "Can`t load ParentConfig: "+parentModelName)
			return nil
		}

		if queryVariableName, exist := config.Parent["queryVariableName"]; exist {
			//fmt.Println("queryVariableName=" + queryVariableName)
			if payload != nil {
				if value, exists := payload[queryVariableName].(string); exists {
					config.Parent["queryVariableValue"] = value
				}
			} else {
				if queryVariableValue, exist := context.GetQuery(queryVariableName); exist {
					//fmt.Println("queryVariableValue=" + queryVariableValue)
					config.Parent["queryVariableValue"] = queryVariableValue
					//			} else {
					//				config.Parent["queryVariableValue"] = ""
				}
			}
		}
	}

	if c.Config.VariableResolver == nil && strings.Contains(config.SqlWhere, "{{") {
		c.somethingWentWrong(context, fmt.Sprintf("Trying to use variables without Config.VariableResolver modelName=%s", modelName))
		return nil
	}

	config.SqlWhere = c.resolveVariables(context, modelName, config.SqlWhere)

	return &config
}

func (c *Impl) loadModelConfigDefaults(config *modelConfig) {
	if config.PageTitle == "" {
		config.PageTitle = config.ModelName
	}
	if config.DbTable == "" {
		config.DbTable = CamelToSnake(config.ModelName)
	}
}

func (c *Impl) fillFieldConfig(config *modelConfig) {
	if config.FieldConfig == nil {
		config.FieldConfig = make(map[string]FieldParams)
	}

	// AddableFields
	for _, field := range config.AddableFields {
		param := config.FieldConfig[field]
		param.IsAddable = true
		if param.FieldEditor == "" {
			param.FieldEditor = "textarea"
		}
		config.FieldConfig[field] = param
	}

	// EditableFields
	for _, field := range config.EditableFields {
		param := config.FieldConfig[field]
		param.IsEditable = true
		if param.FieldEditor == "" {
			param.FieldEditor = "textarea"
		}
		config.FieldConfig[field] = param
	}

	// Required
	for _, field := range config.RequiredFields {
		param := config.FieldConfig[field]
		param.IsRequired = true
		config.FieldConfig[field] = param
	}

	// Classes
	for field, class := range config.Classes {
		param := config.FieldConfig[field]
		param.Classes = class
		config.FieldConfig[field] = param
	}
}

func (c *Impl) resolveVariables(context *gin.Context, modelName string, str string) string {
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

	//queryParams := context.Request.URL.Query()
	for _, variable := range variables {
		value := ""
		//if queryValue, exists := queryParams[variable]; exists {
		//	value = queryValue[0]
		//} else {
		if c.Config.VariableResolver != nil {
			value = c.Config.VariableResolver(context, modelName, variable)
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
