package wedyta

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"strings"
)

func (c *Impl) renderRecordValue(db *gorm.DB, config *modelConfig, field string, record map[string]interface{}, cache *RenderTableCache) (interface{}, string) {
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
		//if cachedValue, found := relatedDataCache[cacheKey]; found {
		if cachedValue, found := cache.RelatedData[cacheKey]; found {
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
						//relatedDataCache[cacheKey] = relatedValue
						cache.RelatedData[cacheKey] = relatedValue
						//log.Printf("Wedyta config problem model=%s relatedData=%s error: %v", modelName, relatedDataField, err)
						log.Printf("WeDyTa: failed to load related value from %s: %v", tableName, err)
					} else {
						value = relatedValue
						//relatedDataCache[cacheKey] = relatedValue
						cache.RelatedData[cacheKey] = relatedValue
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

	return value, tagAttrs
}
