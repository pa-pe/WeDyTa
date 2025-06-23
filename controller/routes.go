package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/embed"
	"github.com/pa-pe/wedyta/service"
	"io/fs"
	"log"
	"net/http"
)

func RegisterRoutes(r *gin.Engine, service *service.Service) {

	//	r.SetHTMLTemplate(loadTemplates())

	staticFiles, err := fs.Sub(embed.EmbeddedFiles, "static")
	if err != nil {
		log.Fatalf("failed to initialize static files: %v", err)
	}

	wedytaGroup := r.Group("/wedyta")
	//wedytaGroup.StaticFS("/static", http.FS(embeddedFiles))
	wedytaGroup.StaticFS("/static", http.FS(staticFiles))
	wedytaGroup.GET("/:modelName", service.RenderTable)
	wedytaGroup.GET("/:modelName/:recID", service.RenderTableRecord)
	wedytaGroup.GET("/:modelName/:recID/:action", func(ctx *gin.Context) { routeModelRecordAction(ctx, service) })
	wedytaGroup.POST("/add", service.HandleTableCreateRecord)
	wedytaGroup.POST("/update", service.Update)
}

func routeModelRecordAction(ctx *gin.Context, s *service.Service) {
	action := ctx.Param("action")
	switch action {
	case "update":
		s.RenderTableRecord(ctx)
	default:
		s.SomethingWentWrong(ctx, "RouteModelRecordAction: unknown action="+action)
	}
}
