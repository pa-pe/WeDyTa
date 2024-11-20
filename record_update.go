package wedyta

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// Update updates the fields of a specified model based on the allowedFields
func (c *Impl) Update(context *gin.Context) {
	var payload map[string]interface{}
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	modelName, ok := payload["modelName"].(string)
	if !ok {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Model name is required"})
		return
	}

	if c.Config.AccessCheckFunc(context, modelName, "", "update") != true {
		//context.String(http.StatusForbidden, "Forbidden RenderTable: "+modelName)
		context.JSON(http.StatusForbidden, gin.H{"error": "Access denied", "modelName": "$modelName"})
		return
	}

	config, err := c.loadModelConfig(context, modelName, payload)
	if err != nil {
		log.Printf("No configuration found for model: %s", modelName)
		context.JSON(http.StatusBadRequest, gin.H{"error": "No configuration found for model '" + modelName + "'"})
		return
	}

	//var payload map[string]interface{}
	//if err := context.ShouldBindJSON(&payload); err != nil {
	//	context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
	//	return
	//}
	//
	//modelName, ok := payload["model"].(string)
	//if !ok {
	//	context.JSON(http.StatusBadRequest, gin.H{"error": "Model name is required"})
	//	return
	//}

	fmt.Println(payload)

	id, ok := payload["id"].(float64)
	if !ok {
		idStr, ok := payload["id"].(string)
		if !ok {
			context.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		idFromStr, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "ID recognize"})
			return
		}
		id = float64(idFromStr)
	}

	var allowed []string

	//Prepare the map for updating
	updateData := make(map[string]interface{})
	//	if class, ok := config.EditableFields[field]; ok {

	//	for _, field := range allowed {
	for field := range config.EditableFields {
		fieldSnaked := CamelToSnake(field)
		if value, exists := payload[fieldSnaked]; exists {
			updateData[fieldSnaked] = value
			allowed = append(allowed, fieldSnaked)
		} else if value, exists := payload[field]; exists {
			updateData[fieldSnaked] = value
			allowed = append(allowed, fieldSnaked)
		}
	}

	if len(updateData) == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	// Retrieve original values for fields to be updated
	originalData := make(map[string]interface{})
	if err := c.DB.Debug().Table(config.DbTable).Where("id = ?", int64(id)).Select(allowed).Take(&originalData).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve original data"})
		return
	}

	if len(updateData) == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	if err := c.DB.Debug().Table(config.DbTable).Where(fmt.Sprint("id = ", id)).Updates(updateData).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update model"})
		return
	}

	if c.Config.AfterUpdate != nil {
		for field, newValue := range updateData {
			originalValue, exists := originalData[field]
			if exists && originalValue != newValue {
				go c.Config.AfterUpdate(context, c.DB, config.DbTable, int64(id), field, fmt.Sprintf("%v", originalValue), fmt.Sprintf("%v", newValue))
			}
		}
	}

	// client side js tests:
	//time.Sleep(1000 * time.Millisecond)
	//context.JSON(http.StatusBadRequest, gin.H{"error": "test fail"})
	//return

	context.JSON(http.StatusOK, gin.H{"success": true, "message": "Model updated successfully"})
}
