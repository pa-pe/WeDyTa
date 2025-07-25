package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"log"
	"strings"
)

func (s *Service) RenderTableRecordCreate(ctx *gin.Context) {
	modelName := ctx.Param("modelName")
	action := "create"

	permit, mConfig := s.checkAccessAndLoadModelConfig(ctx, modelName, action)
	if !permit {
		return
	}

	htmlTable, err := s.RenderModelTableRecordCreate(ctx, mConfig)
	if err != nil {
		s.SomethingWentWrong(ctx, fmt.Sprintf("RenderModelTableRecord error: %v", err))
		return
	}

	s.RenderPage(ctx, mConfig, htmlTable)
}

func (s *Service) RenderModelTableRecordCreate(ctx *gin.Context, mConfig *model.ConfigOfModel) (string, error) {
	if mConfig == nil {
		log.Fatalf("Wedyta: RenderModelTableRecord(): mConfig == nil")
	}

	var htmlTable strings.Builder
	htmlTable.WriteString(s.RenderAddForm(ctx, mConfig))

	return htmlTable.String(), nil
}
