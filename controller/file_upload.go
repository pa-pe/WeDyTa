package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"net/http"
)

func (c *Controller) handleUploadCheck(ctx *gin.Context) {
	var req model.UploadCheckRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, model.UploadCheckResponse{
			Allowed: false,
			Message: "Invalid input.",
		})
		return
	}

	res, err := c.Service.CheckUploadPermission(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.UploadCheckResponse{
			Allowed: false,
			Message: "Server error: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) HandleImageUpload(ctx *gin.Context) {
	imageURL, err := c.Service.ProcessImageUpload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.String(http.StatusOK, imageURL)
}
