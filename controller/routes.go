package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/embed"
	"io/fs"
	"log"
	"net/http"
)

func (c *Controller) RegisterRoutes(r *gin.Engine) {
	s := c.Service
	//	r.SetHTMLTemplate(loadTemplates())

	if c.Service.Config.FileUploadFolder != "" && c.Service.Config.FileUploadRelativePath != "" {
		r.Static(c.Service.Config.FileUploadRelativePath, c.Service.Config.FileUploadFolder)
		c.Service.UploadsConfigured = true
	}

	staticFiles, err := fs.Sub(embed.EmbeddedFiles, "static")
	if err != nil {
		log.Fatalf("failed to initialize static files: %v", err)
	}

	wedytaGroup := r.Group("/wedyta")
	//wedytaGroup.StaticFS("/static", http.FS(embeddedFiles))
	wedytaGroup.StaticFS("/static", http.FS(staticFiles))
	wedytaGroup.GET("/:modelName", s.RenderTable)
	wedytaGroup.GET("/:modelName/create", s.RenderTableRecordCreate)
	wedytaGroup.GET("/:modelName/:recID", s.RenderTableRecord)
	wedytaGroup.GET("/:modelName/:recID/:action", c.routeModelRecordAction)
	wedytaGroup.POST("/add", s.HandleTableCreateRecord)
	wedytaGroup.POST("/update", s.Update)
	wedytaGroup.POST("/upload/check", c.handleUploadCheck)
	wedytaGroup.POST("/upload/image", c.HandleImageUpload)
}

func (c *Controller) routeModelRecordAction(ctx *gin.Context) {
	s := c.Service
	action := ctx.Param("action")

	switch action {
	case "update":
		s.RenderTableRecord(ctx)
	default:
		s.SomethingWentWrong(ctx, "RouteModelRecordAction: unknown action="+action)
	}
}
