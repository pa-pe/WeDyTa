package service

import (
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"gorm.io/gorm"
)

type Service struct {
	DB                *gorm.DB
	Config            *model.WedytaConfig
	modelCache        map[string]model.CachedModelConfig
	UploadsConfigured bool
}

func NewService(db *gorm.DB, wedytaConfig *model.WedytaConfig) *Service {
	if wedytaConfig == nil {
		// default if no wedytaConfig
		wedytaConfig = &model.WedytaConfig{
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

	if wedytaConfig.DebugSQL {
		db = db.Debug()
	}

	if wedytaConfig.JQueryScriptTag == "" {
		wedytaConfig.JQueryScriptTag = "<script src=\"https://code.jquery.com/jquery-3.7.1.min.js\"></script>"
	}

	return &Service{
		DB:                db,
		Config:            wedytaConfig,
		modelCache:        make(map[string]model.CachedModelConfig),
		UploadsConfigured: false,
	}
}
