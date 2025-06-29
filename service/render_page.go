package service

import (
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/embed"
	"github.com/pa-pe/wedyta/model"
	"html/template"
	"net/http"
	"strings"
)

func (s *Service) RenderPage(ctx *gin.Context, mConfig *model.ConfigOfModel, htmlContent string) {
	if s.Config.Template != "" {
		ginH := gin.H{
			"HeaderTags": template.HTML(mConfig.HeaderTags),
			"Title":      mConfig.PageTitle,
			"Content":    template.HTML(htmlContent),
		}
		//ginH["Title"] = mConfig.PageTitle

		if s.Config.PrepareTemplateVariables != nil {
			s.Config.PrepareTemplateVariables(ctx, mConfig.ModelName, ginH)
		}

		ctx.HTML(http.StatusOK, s.Config.Template, ginH)
	} else {
		defaultTemplate := "templates/default.tmpl"
		content, err := embed.EmbeddedFiles.ReadFile(defaultTemplate)
		if err != nil {
			s.SomethingWentWrong(ctx, "Failed to load default template: "+defaultTemplate)
			return
		}

		templateContent := string(content)

		templateContent = strings.Replace(templateContent, "{{ .HeaderTags }}", mConfig.HeaderTags, -1)
		templateContent = strings.Replace(templateContent, "{{ .Title }}", mConfig.PageTitle, -1)
		templateContent = strings.Replace(templateContent, "{{ .Content }}", htmlContent, -1)

		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(templateContent))
	}
}
