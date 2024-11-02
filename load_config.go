package wedyta

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"strings"
)

func (c *RenderTableImpl) loadModelConfig(context *gin.Context, modelName string, payload map[string]interface{}) (*modelConfig, error) {
	configPath := "config/renderModelTable/" + modelName + ".json"

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
			log.Print("Can`t load ParentConfig: " + parentModelName)
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

	queryParams := context.Request.URL.Query()
	for key, values := range queryParams {
		for _, val := range values {
			//		fmt.Printf("Parameter: %s, Value: %s\n", key, value)
			placeholder := fmt.Sprintf("$%s$", key)
			config.SqlWhere = strings.ReplaceAll(config.SqlWhere, placeholder, fmt.Sprintf("%v", val))
		}
	}

	return &config, nil
}
