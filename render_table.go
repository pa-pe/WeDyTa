package wedyta

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/http"
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

	var records []map[string]interface{}
	if err := db.Debug().Table(config.DbTable).Where(config.SqlWhere).Order(config.OrderBy).Find(&records).Error; err != nil {
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
					if num, ok := value.(int64); ok && num != 0 {
						var relatedValue string
						err := db.Debug().Table(strings.Split(relatedDataField, ".")[0]).
							Select(strings.Split(relatedDataField, ".")[1]).
							Where("id = ?", value).
							Row().Scan(&relatedValue)
						if err != nil {
							//return "", err
							relatedDataCache[cacheKey] = relatedValue
							log.Printf("Wedyta config problem model=%s relatedData=%s error: %v", modelName, relatedDataField, err)
						} else {
							value = relatedValue
							relatedDataCache[cacheKey] = relatedValue
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

	return htmlTable.String(), nil
}
