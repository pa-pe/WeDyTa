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

func (c *Impl) RenderTable(ctx *gin.Context) {
	modelName := ctx.Param("modelName")

	if c.Config.AccessCheckFunc(ctx, modelName, "", "read") != true {
		ctx.String(http.StatusForbidden, "No access RenderTable: "+modelName)
		return
	}

	mConfig := c.loadModelConfig(ctx, modelName, nil)
	if mConfig == nil {
		return
	}

	htmlTable, err := c.RenderModelTable(ctx, c.DB, mConfig)
	if err != nil {
		c.somethingWentWrong(ctx, fmt.Sprintf("RenderModelTable error: %v", err))
		return
	}

	if c.Config.Template != "" {
		ginH := gin.H{
			"Title":   mConfig.PageTitle,
			"Content": template.HTML(htmlTable),
		}
		//ginH["Title"] = mConfig.PageTitle

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

		templateContent = strings.Replace(templateContent, "{{ .Title }}", mConfig.PageTitle, -1)
		templateContent = strings.Replace(templateContent, "{{ .Content }}", htmlTable, -1)

		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(templateContent))
	}
}

func (c *Impl) RenderModelTable(ctx *gin.Context, db *gorm.DB, mConfig *modelConfig) (string, error) {
	if mConfig == nil {
		log.Fatalf("Wedyta: RenderModelTable(): mConfig == nil")
	}

	pageNumStr := ctx.Query("page")
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	offset := (pageNum - 1) * c.Config.PaginationRecordsPerPage

	totalRecords, err := c.getTotalRecords(mConfig)
	if err != nil {
		return "", err
	}

	var records []map[string]interface{}
	if err := db.
		Table(mConfig.DbTable).
		Where(mConfig.SqlWhere).
		Order(mConfig.OrderBy).
		Limit(c.Config.PaginationRecordsPerPage).
		Offset(offset).
		Find(&records).Error; err != nil {
		return "", err
	}

	var htmlTable strings.Builder
	htmlTable.WriteString(`<link rel="stylesheet" href="/wedyta/static/css/wedyta.css">` + "\n")

	if len(mConfig.EditableFields) > 0 {
		htmlTable.WriteString(`
<script src="https://code.jquery.com/jquery-3.7.1.min.js"></script>
<script src="/wedyta/static/js/wedyta_update.js"></script>
`)
	}

	htmlTable.WriteString(`<` + c.Config.HeadersTag + `>` + mConfig.PageTitle + `</` + c.Config.HeadersTag + `>` + "\n")
	htmlTable.WriteString(c.breadcrumbBuilder(mConfig, ""))
	htmlTable.WriteString(c.RenderAddForm(ctx, mConfig))

	htmlTable.WriteString("<table class='table table-striped mt-3' model='" + mConfig.ModelName + "'>\n<thead>\n<tr>\n")

	for _, field := range mConfig.Fields {
		header := mConfig.Headers[field]
		if header == "" {
			header = mConfig.Headers[InvertCaseStyle(field)]
		}
		if header == "" {
			header = field
		}

		titleStr := ""
		if title, ok := mConfig.Titles[field]; ok {
			titleStr = fmt.Sprintf(" title='%s'", title)
		}

		htmlTable.WriteString(fmt.Sprintf("<th%s>%s</th>\n", titleStr, header))
	}
	htmlTable.WriteString("</tr>\n</thead>\n<tbody>\n")

	//relatedDataCache := make(map[string]string)
	var cache RenderTableCache
	cache.RelatedData = make(map[string]string)

	for _, record := range records {
		trClass := ""
		if recordIsDisabled, exists := record["is_disabled"]; exists {
			if fmt.Sprint(recordIsDisabled) == "1" {
				trClass = ` class="disabled"`
			}
		}

		htmlTable.WriteString("<tr" + trClass + ">\n")
		for _, field := range mConfig.Fields {
			value, tagAttrs := c.renderRecordValue(mConfig, field, record, &cache)
			htmlTable.WriteString(fmt.Sprintf("\t<td%s>%v</td>\n", tagAttrs, value))
		}
		htmlTable.WriteString("</tr>\n")
	}
	htmlTable.WriteString("</tbody>\n</table>")

	curPageUrl := mConfig.ModelName
	htmlTable.WriteString(c.buildPagination(totalRecords, c.Config.PaginationRecordsPerPage, pageNum, curPageUrl))

	return htmlTable.String(), nil
}

func (c *Impl) buildPagination(totalRecords int64, pageSize int, pageNum int, url string) string {
	pageCount := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))
	if pageCount < 2 {
		return ""
	}

	const delta = 5

	start := pageNum - delta
	if start < 1 {
		start = 1
	}
	end := pageNum + delta
	if end > pageCount {
		end = pageCount
	}

	pagination := "<nav aria-label=\"Page navigation\">\n<ul class=\"pagination justify-content-center\">\n"

	// ← First page
	if start > 1 {
		pagination += fmt.Sprintf("<li class=\"page-item\"><a class=\"page-link\" href=\"%s\">1</a></li>\n", url)
		if start > 2 {
			pagination += "<li class=\"page-item disabled\"><span class=\"page-link\">...</span></li>\n"
		}
	}

	// ← Pages around current
	for i := start; i <= end; i++ {
		url_ := url
		if i > 1 {
			url_ += fmt.Sprintf("?page=%d", i)
		}
		active := ""
		if i == pageNum {
			active = " active"
		}
		pagination += fmt.Sprintf("<li class=\"page-item%s\"><a class=\"page-link\" href=\"%s\">%d</a></li>\n", active, url_, i)
	}

	// → Last page
	if end < pageCount {
		if end < pageCount-1 {
			pagination += "<li class=\"page-item disabled\"><span class=\"page-link\">...</span></li>\n"
		}
		url_ := url + fmt.Sprintf("?page=%d", pageCount)
		pagination += fmt.Sprintf("<li class=\"page-item\"><a class=\"page-link\" href=\"%s\">%d</a></li>\n", url_, pageCount)
	}

	pagination += "</ul>\n</nav>\n"
	return pagination
}
