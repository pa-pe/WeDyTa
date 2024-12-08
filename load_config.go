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

func (c *Impl) loadModelConfig(context *gin.Context, modelName string, payload map[string]interface{}) (*modelConfig, error) {
	configPath := c.Config.ConfigDir + "/" + modelName + ".json"

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config modelConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	if config.PageTitle == "" {
		config.PageTitle = modelName
	}
	if config.DbTable == "" {
		config.DbTable = CamelToSnake(modelName)
	}

	if parentModelName, parentExists := config.Parent["modelName"]; parentExists {
		//fmt.Println(parentModelName)
		config.ParentConfig, err = c.loadModelConfig(context, parentModelName, payload)
		if err != nil {
			log.Print("Wedyta: Can`t load ParentConfig: " + parentModelName)
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
		return nil, fmt.Errorf("trying to use variables without Config.VariableResolver modelName=%s", modelName)
	}

	config.SqlWhere = c.resolveVariables(context, modelName, config.SqlWhere)

	return &config, nil
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
