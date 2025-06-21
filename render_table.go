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
	"time"
)

func (c *Impl) RenderTable(context *gin.Context) {
	modelName := context.Param("modelName")

	if c.Config.AccessCheckFunc(context, modelName, "", "read") != true {
		context.String(http.StatusForbidden, "No access RenderTable: "+modelName)
		return
	}

	config := c.loadModelConfig(context, modelName, nil)
	if config == nil {
		return
	}

	htmlTable, err := c.RenderModelTable(context, c.DB, modelName, config)
	if err != nil {
		c.somethingWentWrong(context, fmt.Sprintf("RenderModelTable error: %v", err))
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

func (c *Impl) RenderModelTable(context *gin.Context, db *gorm.DB, modelName string, config *modelConfig) (string, error) {
	if config == nil || modelName == "" {
		log.Fatalf("configuration not found for model: %s", modelName)
	}

	pageNumStr := context.Query("page")
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	offset := (pageNum - 1) * c.Config.PaginationRecordsPerPage

	totalRecords, err := c.getTotalRecords(db, config)
	if err != nil {
		return "", err
	}

	var records []map[string]interface{}
	if err := db.Debug().
		Table(config.DbTable).
		Where(config.SqlWhere).
		Order(config.OrderBy).
		Limit(c.Config.PaginationRecordsPerPage).
		Offset(offset).
		Find(&records).Error; err != nil {
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

	htmlTable.WriteString(`<` + c.Config.HeadersTag + `>` + config.PageTitle + `</` + c.Config.HeadersTag + `>` + "\n")
	htmlTable.WriteString(c.breadcrumbBuilder(config))
	htmlTable.WriteString(c.RenderAddForm(context, config, modelName))

	htmlTable.WriteString("<table class='table table-striped mt-3' model='" + modelName + "'>\n<thead>\n<tr>\n")

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

		htmlTable.WriteString(fmt.Sprintf("<th%s>%s</th>\n", titleStr, header))
	}
	htmlTable.WriteString("</tr>\n</thead>\n<tbody>\n")

	relatedDataCache := make(map[string]string)

	for _, record := range records {
		trClass := ""
		if recordIsDisabled, exists := record["is_disabled"]; exists {
			if fmt.Sprint(recordIsDisabled) == "1" {
				trClass = ` class="disabled"`
			}
		}

		htmlTable.WriteString("<tr" + trClass + ">\n")
		for _, field := range config.Fields {
			//value := record[field]
			value, exists := record[field]
			if !exists || value == nil {
				value, exists = record[InvertCaseStyle(field)]
				if !exists || value == nil {
					value = ""
				}
			}

			fldCfg := config.FieldConfig[field]

			classStr := ""
			additionalAttr := ""
			if field == "id" || field == "ID" {
				classStr += " rec_id"
			}

			if fldCfg.Classes != "" {
				classStr += " " + fldCfg.Classes
			}

			if fldCfg.IsEditable {
				classStr += " editable editable-" + fldCfg.FieldEditor
				additionalAttr += ` fieldName="` + CamelToSnake(field) + `"`
			}

			classAttr := ""
			if classStr != "" {
				classStr = classStr[1:]
				classAttr = fmt.Sprintf(" class='%s'", classStr)
			}
			tagAttrs := classAttr + additionalAttr

			relatedDataField, relatedExists := config.RelatedData[field]
			if relatedExists {
				//relatedDataField = utils.CamelToSnake(relatedDataField)
				cacheKey := fmt.Sprintf("%s_%v", relatedDataField, value)
				if cachedValue, found := relatedDataCache[cacheKey]; found {
					value = cachedValue
				} else {
					num := extractInt64(value)

					if num != 0 {
						tableParts := strings.Split(relatedDataField, ".")
						tableName := tableParts[0]
						fieldName := tableParts[1]

						pkField, err := getPrimaryKeyFieldName(db, tableName)
						if err != nil {
							log.Printf("WeDyTa: can't determine primary key for table %s: %v", tableName, err)
						} else {
							var relatedValue string
							err = db.Debug().
								Table(tableName).
								Select(fieldName).
								Where(fmt.Sprintf("%s = ?", pkField), value).
								Row().
								Scan(&relatedValue)

							if err != nil {
								//	return "", err
								relatedDataCache[cacheKey] = relatedValue
								//log.Printf("Wedyta config problem model=%s relatedData=%s error: %v", modelName, relatedDataField, err)
								log.Printf("WeDyTa: failed to load related value from %s: %v", tableName, err)
							} else {
								value = relatedValue
								relatedDataCache[cacheKey] = relatedValue
							}
						}
					}
				}
			}

			if countConfig, countExists := config.CountRelatedData[field]; countExists {
				foreignKeyValue, ok := record[countConfig.LocalFieldID]
				var count int64
				if ok {
					if err := db.Debug().Table(countConfig.Table).
						Where(fmt.Sprintf("%s = ?", countConfig.TargetFieldID), foreignKeyValue).
						Count(&count).Error; err == nil {
					}
				}
				value = count
			}

			if linkConfig, linkExists := config.Links[field]; linkExists {
				link := linkConfig.Template
				for key, val := range record {
					placeholder := fmt.Sprintf("$%s$", key)
					link = strings.ReplaceAll(link, placeholder, fmt.Sprintf("%v", val))
				}
				value = fmt.Sprintf("<a href='%s'>%v</a>", link, value)
			}

			if dateTimeFieldConfig, dateTimeFieldExists := config.DateTimeFields[field]; dateTimeFieldExists {
				value = extractFormattedTime(value, dateTimeFieldConfig)
			}

			htmlTable.WriteString(fmt.Sprintf("\t<td%s>%v</td>\n", tagAttrs, value))
		}
		htmlTable.WriteString("</tr>\n")
	}
	htmlTable.WriteString("</tbody>\n</table>")

	curPageUrl := modelName
	htmlTable.WriteString(c.buildPagination(totalRecords, c.Config.PaginationRecordsPerPage, pageNum, curPageUrl))

	return htmlTable.String(), nil
}

func (c *Impl) getTotalRecords(db *gorm.DB, config *modelConfig) (int64, error) {
	var totalRecords int64
	if err := db.Debug().Table(config.DbTable).Where(config.SqlWhere).Count(&totalRecords).Error; err != nil {
		return 0, err
	}

	return totalRecords, nil
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

func extractInt64(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int8:
		return int64(v)
	case uint8:
		return int64(v)
	case int32:
		return int64(v)
	case uint64:
		return int64(v)
	case uint32:
		return int64(v)
	case int:
		return int64(v)
	case uint:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case []byte:
		if num, err := strconv.ParseInt(string(v), 10, 64); err == nil {
			return num
		}
	case string:
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			return num
		}
	default:
		log.Printf("WeDyTa TODO: unhandled type %T with value %v", value, value)
	}
	return 0
}

var primaryKeyCache = make(map[string]string)

func getPrimaryKeyFieldName(db *gorm.DB, tableName string) (string, error) {
	if pk, ok := primaryKeyCache[tableName]; ok {
		return pk, nil
	}

	var columnName string
	query := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND COLUMN_KEY = 'PRI'
		LIMIT 1
	`
	err := db.Raw(query, tableName).Scan(&columnName).Error
	if err != nil {
		return "", err
	}

	primaryKeyCache[tableName] = columnName

	return columnName, nil
}

func extractFormattedTime(value any, outputFormat string) string {
	switch v := value.(type) {
	case time.Time:
		return v.Format(outputFormat)
	case string:
		for _, layout := range []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
			"02.01.2006 15:04:05",
			"02.01.2006",
		} {
			if t, err := time.Parse(layout, v); err == nil {
				return t.Format(outputFormat)
			}
		}
		return v // fallback
	case []byte:
		return extractFormattedTime(string(v), outputFormat)
	default:
		return fmt.Sprintf("%v", value)
	}
}
