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
	DB         *gorm.DB
	Config     *Config
	modelCache map[string]cachedModelConfig
}

func New(r *gin.Engine, db *gorm.DB, wedytaConfig *Config) *Impl {
	if wedytaConfig == nil {
		// default if no wedytaConfig
		wedytaConfig = &Config{
			AccessCheckFunc: func(context *gin.Context, modelName, action, fieldName string) bool {
				return true // default permit all
			},
		}
	}

	if wedytaConfig.ConfigDir == "" {
		wedytaConfig.ConfigDir = "config/wedyta"
	}

	if wedytaConfig.HeadersTag == "" {
		wedytaConfig.HeadersTag = "h2"
	}

	if wedytaConfig.PaginationRecordsPerPage == 0 {
		wedytaConfig.PaginationRecordsPerPage = 100
	}

	if wedytaConfig.BreadcrumbsRootName == "" {
		wedytaConfig.BreadcrumbsRootName = "Home"
	}

	if wedytaConfig.BreadcrumbsRootUrl == "" {
		wedytaConfig.BreadcrumbsRootUrl = "/"
	}

	if wedytaConfig.BreadcrumbsDivider == "" {
		wedytaConfig.BreadcrumbsDivider = ">"
	}

	db_ := db
	if wedytaConfig.DebugSQL {
		db_ = db.Debug()
	}

	impl := &Impl{
		DB:     db_,
		Config: wedytaConfig,
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
	wedytaGroup.GET("/:modelName/:recID/:action", impl.routeModelRecordAction)
	wedytaGroup.POST("/add", impl.HandleTableCreateRecord)
	wedytaGroup.POST("/update", impl.Update)

	return impl
}

func (c *Impl) routeModelRecordAction(ctx *gin.Context) {
	action := ctx.Param("action")
	switch action {
	case "update":
		c.RenderTableRecord(ctx)
	default:
		c.somethingWentWrong(ctx, "RouteModelRecordAction: unknown action="+action)
	}
}
