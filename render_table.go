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

func (c *Impl) RenderTable(context *gin.Context) {
	modelName := context.Param("modelName")

	if c.Config.AccessCheckFunc(context, modelName, "", "read") != true {
		context.String(http.StatusForbidden, "No access RenderTable: "+modelName)
		return
	}

	config, err := c.loadModelConfig(context, modelName, nil)
	if err != nil {
		log.Print("No configuration found for RenderTable: " + modelName)
		context.String(http.StatusNotFound, "No configuration found for RenderTable: "+modelName)
		return
	}

	htmlTable, err := c.RenderModelTable(context, c.DB, modelName, config)
	if err != nil {
		fmt.Println("Ошибка:", err)
		errStr := fmt.Sprint("Ошибка:", err)
		context.String(http.StatusNotFound, "Error RenderTable "+modelName+": "+errStr)
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
		content, err := embeddedFiles.ReadFile("templates/default.tmpl")
		if err != nil {
			context.String(http.StatusInternalServerError, "Failed to load vedyta default template")
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

			classStr := ""
			additionalAttr := ""
			if field == "id" || field == "ID" {
				classStr += " rec_id"
			}

			if class, ok := config.Classes[field]; ok {
				//classAttr = fmt.Sprintf(" class='%s'", class)
				classStr += " " + class
			}

			if class, ok := config.EditableFields[field]; ok {
				classStr += " editable editable-" + class
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

	pagination := "<nav aria-label=\"Page navigation\">\n<ul class=\"pagination justify-content-center\">\n"
	for i := 1; i <= pageCount; i++ {
		url_ := url
		if i > 1 {
			url_ += fmt.Sprintf("?page=%d", i)
		}

		active := ""
		if i == pageNum {
			active = " active"
		}

		pagination += "<li class=\"page-item" + active + "\">"
		pagination += fmt.Sprintf("<a class=\"page-link\" href=\"%s\">%d</a> ", url_, i)
		pagination += "</li>"
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
