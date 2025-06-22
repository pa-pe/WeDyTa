package wedyta

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// Update updates the fields of a specified model based on the allowedFields
func (c *Impl) Update(ctx *gin.Context) {
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

	if c.Config.AccessCheckFunc(ctx, modelName, "", "update") != true {
		//ctx.String(http.StatusForbidden, "Forbidden RenderTable: "+modelName)
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Access denied", "modelName": "$modelName"})
		return
	}

	config := c.loadModelConfig(ctx, modelName, payload)
	if config == nil {
		return
	}

	//var payload map[string]interface{}
	//if err := ctx.ShouldBindJSON(&payload); err != nil {
	//	ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
	//	return
	//}
	//
	//modelName, ok := payload["model"].(string)
	//if !ok {
	//	ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model name is required"})
	//	return
	//}

	fmt.Println(payload)

	id, ok := payload["id"].(float64)
	if !ok {
		idStr, ok := payload["id"].(string)
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		idFromStr, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID recognize"})
			return
		}
		id = float64(idFromStr)
	}

	var allowed []string

	//Prepare the map for updating
	updateData := make(map[string]interface{})
	//	if class, ok := config.EditableFields[field]; ok {

	//for field := range config.EditableFields {
	for _, field := range config.EditableFields {
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	// Retrieve original values for fields to be updated
	originalData := make(map[string]interface{})
	if err := c.DB.Debug().Table(config.DbTable).Where("id = ?", int64(id)).Select(allowed).Take(&originalData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve original data"})
		return
	}

	if len(updateData) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	if err := c.DB.Debug().Table(config.DbTable).Where(fmt.Sprint("id = ", id)).Updates(updateData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update model"})
		return
	}

	if c.Config.AfterUpdate != nil {
		for field, newValue := range updateData {
			originalValue, exists := originalData[field]
			if exists && originalValue != newValue {
				go c.Config.AfterUpdate(ctx, c.DB, config.DbTable, int64(id), field, fmt.Sprintf("%v", originalValue), fmt.Sprintf("%v", newValue))
			}
		}
	}

	// client side js tests:
	//time.Sleep(1000 * time.Millisecond)
	//ctx.JSON(http.StatusBadRequest, gin.H{"error": "test fail"})
	//return

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Model updated successfully"})
}
