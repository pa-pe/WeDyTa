package wedyta

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (c *Impl) RenderTableRecord(ctx *gin.Context) {
	modelName := ctx.Param("modelName")
	recIDstr := ctx.Param("recID")
	recID, err := strconv.ParseInt(recIDstr, 10, 64)
	if err != nil {
		c.somethingWentWrong(ctx, "Can't ParseInt recID="+recIDstr)
	}

	action := ctx.Param("action")
	isUpdateMode := false
	if action == "" {
		action = "read"
	} else if action == "update" {
		isUpdateMode = true
	} else {
		c.somethingWentWrong(ctx, "Can't ParseInt action="+action)
	}

	if c.Config.AccessCheckFunc(ctx, modelName, "", action) != true {
		ctx.String(http.StatusForbidden, "No access RenderTable: "+modelName)
		return
	}

	config := c.loadModelConfig(ctx, modelName, nil)
	if config == nil {
		return
	}

	htmlTable, err := c.RenderModelTableRecord(ctx, modelName, config, recID, isUpdateMode)
	if err != nil {
		c.somethingWentWrong(ctx, fmt.Sprintf("RenderModelTableRecord error: %v", err))
		return
	}

	if c.Config.Template != "" {
		ginH := gin.H{
			"Title":   config.PageTitle,
			"Content": template.HTML(htmlTable),
		}
		//ginH["Title"] = config.PageTitle

		if c.Config.PrepareTemplateVariables != nil {
			c.Config.PrepareTemplateVariables(ctx, modelName, ginH)
		}

		ctx.HTML(http.StatusOK, c.Config.Template, ginH)
	} else {
		defaultTemplate := "templates/default.tmpl"
		content, err := embeddedFiles.ReadFile(defaultTemplate)
		if err != nil {
			c.somethingWentWrong(ctx, "Failed to load default template: "+defaultTemplate)
			return
		}

		templateContent := string(content)

		templateContent = strings.Replace(templateContent, "{{ .Title }}", config.PageTitle, -1)
		templateContent = strings.Replace(templateContent, "{{ .Content }}", htmlTable, -1)

		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(templateContent))
	}
}

func (c *Impl) RenderModelTableRecord(ctx *gin.Context, modelName string, config *modelConfig, recID int64, isUpdateMode bool) (string, error) {
	if config == nil || modelName == "" {
		log.Fatalf("configuration not found for model: %s", modelName)
	}

	if config.DbTablePrimaryKey == "" {
		err := fmt.Errorf("empty config.DbTablePrimaryKey for model: %s", modelName)
		return "", err
	}

	var record map[string]interface{}
	if err := c.DB.
		Model(&record).
		Table(config.DbTable).
		Where(fmt.Sprintf("%s = %d", config.DbTablePrimaryKey, recID)).
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

	htmlTable.WriteString(`
<style>
table { width: auto !important; }
th { white-space: nowrap; width: 55px; }
td { width: auto !important; }
</style>
`)

	htmlTable.WriteString("<div class=\"col\">\n")

	htmlTable.WriteString(`<` + c.Config.HeadersTag + `>` + config.PageTitle + `</` + c.Config.HeadersTag + `>` + "\n")
	htmlTable.WriteString(c.breadcrumbBuilder(config, fmt.Sprintf("%d", recID)))

	if isUpdateMode {
		var pkValue string
		value, exists := record[config.DbTablePrimaryKey]
		if exists {
			pkValue = fmt.Sprintf("%v", value)
		} else {
			c.somethingWentWrong(ctx, "Can't take primary key value")
		}

		htmlTable.WriteString("<form id=\"editForm\">\n")
		htmlTable.WriteString(" <input type=\"hidden\" name=\"modelName\" value=\"" + config.ModelName + "\">\n")
		htmlTable.WriteString("<input type=\"hidden\" name=\"id\" value=\"" + pkValue + "\">\n")
	}

	htmlTable.WriteString("<table class='table table-striped mt-3' model='" + modelName + "'>\n<tbody>\n<tr>\n")

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
		fldCfg := config.FieldConfig[field]

		value, tagAttrs := c.renderRecordValue(config, field, record, &cache)
		if isUpdateMode && fldCfg.IsEditable {
			htmlTable.WriteString(fmt.Sprintf("<tr>\n <td%s colspan=\"2\"><span%s>%s:</span><br>\n<textarea class=\"form-control\" name=\"%s\">%v</textarea></td>\n</tr>\n", tagAttrs, titleStr, header, field, value))
		} else {
			htmlTable.WriteString(fmt.Sprintf("<tr>\n <th%s>%s:</th>\n <td%s>%v</td>\n</tr>\n", titleStr, header, tagAttrs, value))
		}
	}
	htmlTable.WriteString("</tbody>\n</table>\n")

	if isUpdateMode {
		htmlTable.WriteString("<button type=\"button\" class=\"btn btn-primary\" id=\"saveButton\">Save</button>\n")
		htmlTable.WriteString("</form>\n")
	}

	htmlTable.WriteString("</div>\n")

	return htmlTable.String(), nil
}
