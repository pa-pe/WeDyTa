package wedyta

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (c *Impl) RenderTableRecord(context *gin.Context) {
	modelName := context.Param("modelName")
	recIDstr := context.Param("recID")
	recID, err := strconv.ParseInt(recIDstr, 10, 64)
	if err != nil {
		c.somethingWentWrong(context, "Can't ParseInt recID="+recIDstr)
	}

	if c.Config.AccessCheckFunc(context, modelName, "", "read") != true {
		context.String(http.StatusForbidden, "No access RenderTable: "+modelName)
		return
	}

	config := c.loadModelConfig(context, modelName, nil)
	if config == nil {
		return
	}

	htmlTable, err := c.RenderModelTableRecord(context, c.DB, modelName, config, recID)
	if err != nil {
		c.somethingWentWrong(context, fmt.Sprintf("RenderModelTableRecord error: %v", err))
		return
	}

	if c.Config.Template != "" {
		ginH := gin.H{
			"Title":   config.PageTitle,
			"Content": template.HTML(htmlTable),
		}
		//ginH["Title"] = config.PageTitle

		if c.Config.PrepareTemplateVariables != nil {
			c.Config.PrepareTemplateVariables(context, modelName, ginH)
		}

		context.HTML(http.StatusOK, c.Config.Template, ginH)
	} else {
		defaultTemplate := "templates/default.tmpl"
		content, err := embeddedFiles.ReadFile(defaultTemplate)
		if err != nil {
			c.somethingWentWrong(context, "Failed to load default template: "+defaultTemplate)
			return
		}

		templateContent := string(content)

		templateContent = strings.Replace(templateContent, "{{ .Title }}", config.PageTitle, -1)
		templateContent = strings.Replace(templateContent, "{{ .Content }}", htmlTable, -1)

		context.Data(http.StatusOK, "text/html; charset=utf-8", []byte(templateContent))
	}
}

func (c *Impl) RenderModelTableRecord(context *gin.Context, db *gorm.DB, modelName string, config *modelConfig, recID int64) (string, error) {
	_ = context
	if config == nil || modelName == "" {
		log.Fatalf("configuration not found for model: %s", modelName)
	}

	pkField, err := getPrimaryKeyFieldName(db, config.DbTable)
	if err != nil {
		return "", err
	}

	var record map[string]interface{}
	if err := db.Debug().
		Model(&record).
		Table(config.DbTable).
		Where(fmt.Sprintf("%s = %d", pkField, recID)).
		Where(config.SqlWhere).
		Take(&record).Error; err != nil {
		return "", err
	}

	var htmlTable strings.Builder
	htmlTable.WriteString(`<link rel="stylesheet" href="/wedyta/static/css/wedyta.css">` + "\n")

	if len(config.EditableFields) > 0 {
		htmlTable.WriteString(`
<script src="https://code.jquery.com/jquery-3.7.1.min.js"></script>
<script src="/wedyta/static/js/wedyta_update.js"></script>
`)
	}

	htmlTable.WriteString("<div class=\"col\">\n")

	htmlTable.WriteString(`<` + c.Config.HeadersTag + `>` + config.PageTitle + `</` + c.Config.HeadersTag + `>` + "\n")
	htmlTable.WriteString(c.breadcrumbBuilder(config, fmt.Sprintf("%d", recID)))

	htmlTable.WriteString("<table class='table table-striped mt-3' model='" + modelName + "' style='width: auto;'>\n<tbody>\n<tr>\n")

	for _, field := range config.Fields {
		header := config.Headers[field]
		if header == "" {
			header = config.Headers[InvertCaseStyle(field)]
		}
		if header == "" {
			header = field
		}

		titleStr := ""
		if title, ok := config.Titles[field]; ok {
			titleStr = fmt.Sprintf(" title='%s'", title)
		}

		var cache RenderTableCache
		cache.RelatedData = make(map[string]string)

		value, tagAttrs := c.renderRecordValue(db, config, field, record, &cache)
		htmlTable.WriteString(fmt.Sprintf("<tr>\n <th%s>%s:</th>\n <td%s>%v</td>\n</tr>\n", titleStr, header, tagAttrs, value))
	}
	htmlTable.WriteString("</tbody>\n</table>\n")
	htmlTable.WriteString("</div>\n")

	return htmlTable.String(), nil
}
