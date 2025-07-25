package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"github.com/pa-pe/wedyta/utils/sqlutils"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
)

func (s *Service) RenderTable(ctx *gin.Context) {
	modelName := ctx.Param("modelName")

	action := "read"
	permit, mConfig := s.checkAccessAndLoadModelConfig(ctx, modelName, action)
	if !permit {
		return
	}

	htmlTable, err := s.RenderModelTable(ctx, s.DB, mConfig)
	if err != nil {
		s.SomethingWentWrong(ctx, fmt.Sprintf("RenderModelTable error: %v", err))
		return
	}

	s.RenderPage(ctx, mConfig, htmlTable)
}

func (s *Service) RenderModelTable(ctx *gin.Context, db *gorm.DB, mConfig *model.ConfigOfModel) (string, error) {
	if mConfig == nil {
		log.Fatalf("Wedyta: RenderModelTable(): mConfig == nil")
	}

	pageNumStr := ctx.Query("page")
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	offset := (pageNum - 1) * s.Config.PaginationRecordsPerPage

	totalRecords, err := sqlutils.GetTotalRecords(s.DB, mConfig)
	if err != nil {
		return "", err
	}

	var records []map[string]interface{}
	if err := db.
		Table(mConfig.DbTable).
		Where(mConfig.SqlWhere).
		Order(mConfig.OrderBy).
		Limit(s.Config.PaginationRecordsPerPage).
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

	htmlTable.WriteString(`<` + s.Config.HeadersTag + `>` + mConfig.PageTitle + `</` + s.Config.HeadersTag + `>` + "\n")
	htmlTable.WriteString(s.breadcrumbBuilder(mConfig, "", "read records"))
	htmlTable.WriteString(s.wrapBsAccordion(s.RenderAddForm(ctx, mConfig), "", "Add New Record"))

	htmlTable.WriteString("<table class='table table-striped mt-3' model='" + mConfig.ModelName + "'>\n<thead>\n<tr>\n")

	for _, field := range mConfig.Fields {
		if !mConfig.FieldConfig[field].PermitDisplayInTableMode {
			continue
		}

		header := mConfig.FieldConfig[field].Header

		titleStr := ""
		if title, ok := mConfig.Titles[field]; ok {
			titleStr = fmt.Sprintf(" title='%s'", title)
		}

		htmlTable.WriteString(fmt.Sprintf("<th%s id=\"header_of_%s\">%s</th>\n", titleStr, field, header))
	}
	htmlTable.WriteString("</tr>\n</thead>\n<tbody>\n")

	//relatedDataCache := make(map[string]string)
	var cache model.RenderTableCache
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
			if !mConfig.FieldConfig[field].PermitDisplayInTableMode {
				continue
			}

			value, tagAttrs := s.renderRecordValue(ctx, mConfig, field, record, &cache)
			htmlTable.WriteString(fmt.Sprintf("\t<td%s>%v</td>\n", tagAttrs, value))
		}
		htmlTable.WriteString("</tr>\n")
	}
	htmlTable.WriteString("</tbody>\n</table>")

	curPageUrl := mConfig.ModelName
	htmlTable.WriteString(s.buildPagination(totalRecords, s.Config.PaginationRecordsPerPage, pageNum, curPageUrl))

	return htmlTable.String(), nil
}

func (s *Service) buildPagination(totalRecords int64, pageSize int, pageNum int, url string) string {
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
