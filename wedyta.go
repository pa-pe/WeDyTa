package wedyta

import (
	"embed"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
)

//go:embed static/* templates/*
var embeddedFiles embed.FS

type Impl struct {
	DB     *gorm.DB
	Config *RenderDbTableConfig
}

func New(r *gin.Engine, db *gorm.DB, config *RenderDbTableConfig) *Impl {
	if config == nil {
		// default if no config
		config = &RenderDbTableConfig{
			AccessCheckFunc: func(context *gin.Context, modelName, action, fieldName string) bool {
				return true // default permit all
			},
		}
	}

	impl := &Impl{
		DB:     db,
		Config: config,
	}

	//r.StaticFS("/wedyta/static", http.FS(embeddedFiles))
	r.SetHTMLTemplate(loadTemplates())

	//r.GET("/wedyta/:modelName", impl.RenderTable)
	//r.POST("/wedyta/add/", impl.HandleRenderTableAddRecord)
	//r.POST("/wedyta/update/", impl.Update)

	wedytaGroup := r.Group("/wedyta")
	wedytaGroup.StaticFS("/static", http.FS(embeddedFiles))
	wedytaGroup.GET("/:modelName", impl.RenderTable)
	wedytaGroup.POST("/add", impl.HandleRenderTableAddRecord)
	wedytaGroup.POST("/update", impl.Update)

	return impl
}

// loadTemplates load builtin HTML-templates
func loadTemplates() *template.Template {
	tmpl := template.New("")
	err := filepath.WalkDir("templates", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			content, _ := embeddedFiles.ReadFile(path)
			_, err := tmpl.New(path).Parse(string(content))
			if err != nil {
				log.Fatalf("failed to parse templates: %v", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}
	return tmpl
}
