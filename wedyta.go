package wedyta

import (
	"embed"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static/* templates/default.tmpl
var embeddedFiles embed.FS

type Impl struct {
	DB     *gorm.DB
	Config *Config
}

func New(r *gin.Engine, db *gorm.DB, config *Config) *Impl {
	if config == nil {
		// default if no config
		config = &Config{
			AccessCheckFunc: func(context *gin.Context, modelName, action, fieldName string) bool {
				return true // default permit all
			},
		}
	}

	if config.ConfigDir == "" {
		config.ConfigDir = "config/wedyta"
	}

	if config.HeadersTag == "" {
		config.HeadersTag = "h2"
	}

	if config.PaginationRecordsPerPage == 0 {
		config.PaginationRecordsPerPage = 100
	}

	if config.BreadcrumbsRootName == "" {
		config.BreadcrumbsRootName = "Home"
	}

	if config.BreadcrumbsRootUrl == "" {
		config.BreadcrumbsRootUrl = "/"
	}

	if config.BreadcrumbsDivider == "" {
		config.BreadcrumbsDivider = ">"
	}

	impl := &Impl{
		DB:     db,
		Config: config,
	}

	//	r.SetHTMLTemplate(loadTemplates())

	staticFiles, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatalf("failed to initialize static files: %v", err)
	}

	wedytaGroup := r.Group("/wedyta")
	//wedytaGroup.StaticFS("/static", http.FS(embeddedFiles))
	wedytaGroup.StaticFS("/static", http.FS(staticFiles))
	wedytaGroup.GET("/:modelName", impl.RenderTable)
	wedytaGroup.GET("/:modelName/:recID", impl.RenderTableRecord)
	wedytaGroup.GET("/:modelName/:recID/:action", impl.RouteModelRecordAction)
	wedytaGroup.POST("/add", impl.HandleTableCreateRecord)
	wedytaGroup.POST("/update", impl.Update)

	return impl
}

func (c *Impl) RouteModelRecordAction(ctx *gin.Context) {
	action := ctx.Param("action")
	switch action {
	case "update":
		c.RenderTableRecord(ctx)
	default:
		c.somethingWentWrong(ctx, "RouteModelRecordAction: unknown action="+action)
	}
}
