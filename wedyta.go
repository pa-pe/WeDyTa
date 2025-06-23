package wedyta

import (
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/controller"
	"github.com/pa-pe/wedyta/model"
	"github.com/pa-pe/wedyta/service"
	"gorm.io/gorm"
)

var _ = New

func New(r *gin.Engine, db *gorm.DB, cfg *model.Config) *service.Service {
	impl := service.NewService(db, cfg)
	controller.RegisterRoutes(r, impl)
	return impl
}
