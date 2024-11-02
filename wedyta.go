package wedyta

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Impl struct {
	DB     *gorm.DB
	Config *RenderDbTableConfig
}

func New(db *gorm.DB, config *RenderDbTableConfig) *Impl {
	if config == nil {
		// default if no config
		config = &RenderDbTableConfig{
			AccessCheckFunc: func(context *gin.Context, modelName, action, fieldName string) bool {
				return true // default permit all
			},
		}
	}

	return &Impl{
		DB:     db,
		Config: config,
	}
}
