package service

import (
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"log"
	"net/http"
)

func (s *Service) checkAccessAndLoadModelConfig(ctx *gin.Context, modelName string, action string) (bool, *model.ConfigOfModel) {
	if s.Config.AccessCheckFunc(ctx, modelName, "", action) != true {
		ctx.String(http.StatusForbidden, "Access Denied")
		return false, nil
	}

	mConfig := s.loadModelConfig(ctx, modelName, nil)
	if mConfig == nil {
		return false, nil
	}

	return true, mConfig
}

func (s *Service) SomethingWentWrong(ctx *gin.Context, logString string) {
	log.Println("Wedyta: " + logString + " url=" + ctx.Request.URL.String())
	ctx.String(http.StatusInternalServerError, "Something went wrong, see log for details.")
}
