package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"github.com/pa-pe/wedyta/utils"
	"github.com/pa-pe/wedyta/utils/sqlutils"
	"log"
	"strings"
)

func takeFieldValueFromRecord(field string, record map[string]interface{}) interface{} {
	value, exists := record[field]
	if !exists || value == nil {
		value, exists = record[utils.InvertCaseStyle(field)]
		if !exists || value == nil {
			value = ""
		}
	}

	return value
}

func (s *Service) renderRecordValue(ctx *gin.Context, mConfig *model.ConfigOfModel, field string, record map[string]interface{}, cache *model.RenderTableCache) (interface{}, string) {
	value := takeFieldValueFromRecord(field, record)

	var pkValue string
	pkValueI, exists := record[mConfig.DbTablePrimaryKey]
	if exists {
		pkValue = fmt.Sprintf("%v", pkValueI)
	}

	fldCfg := mConfig.FieldConfig[field]

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
		additionalAttr += ` fieldName="` + utils.CamelToSnake(field) + `"`
	}

	classAttr := ""
	if classStr != "" {
		classStr = classStr[1:]
		classAttr = fmt.Sprintf(" class='%s'", classStr)
	}
	tagAttrs := classAttr + additionalAttr

	columnDataFunc, exists := mConfig.ColumnDataFunc[field]
	if exists {
		if columnDataFunc == "stdRecordControls" {
			url := "/wedyta/" + mConfig.ModelName + "/" + pkValue + "/update" + mConfig.AdditionalUrlParams
			value = "<a href=\"" + url + "\"><i class=\"bi-pen record-control-update\"></i></a>"
			//value = value.(string) + " <i class=\"bi-trash record-control-delete\"></i>"
		} else if columnDataFunc == "dynamicColumnDataFunc" {
			if s.Config.DynamicColumnDataFunc != nil {
				value = s.Config.DynamicColumnDataFunc(ctx, s.DB, mConfig.DbTable, field, record)
			} else {
				value = "!WedytaConfig.DynamicColumnDataFunc not set"
			}
		} else {
			value = "!Unknown columnDataFunc: " + columnDataFunc
		}
	}

	if fldCfg.RelatedData != nil {
		rdCfg := fldCfg.RelatedData
		cacheKey := fmt.Sprintf("%s_%v", rdCfg.TableAndField, value)
		if cachedValue, found := cache.RelatedData[cacheKey]; found {
			value = cachedValue
		} else {
			num := sqlutils.ExtractInt64(value)

			if num != 0 {
				var relatedValue string
				err := s.DB.
					Table(rdCfg.TableName).
					Select(rdCfg.FieldName).
					Where(fmt.Sprintf("%s = ?", rdCfg.PrimaryKeyFieldName), value).
					Row().
					Scan(&relatedValue)

				if err != nil {
					log.Printf("WeDyTa: failed to load related value from %s %s=%v err: %v", fldCfg.RelatedData.TableName, rdCfg.PrimaryKeyFieldName, value, err)
					relatedValue = fmt.Sprintf("#%v", value)
					value = relatedValue
					cache.RelatedData[cacheKey] = relatedValue
				} else {
					value = relatedValue
					cache.RelatedData[cacheKey] = relatedValue
				}
			}
		}
	}

	if countConfig, countExists := mConfig.CountRelatedData[field]; countExists {
		foreignKeyValue, ok := record[countConfig.LocalFieldID]
		var count int64
		if ok {
			if err := s.DB.Table(countConfig.Table).
				Where(fmt.Sprintf("%s = ?", countConfig.TargetFieldID), foreignKeyValue).
				Count(&count).Error; err == nil {
			}
		}
		value = count
	}

	if linkConfig, linkExists := mConfig.Links[field]; linkExists {
		link := linkConfig.Template
		for key, val := range record {
			placeholder := fmt.Sprintf("$%s$", key)
			link = strings.ReplaceAll(link, placeholder, fmt.Sprintf("%v", val))
		}
		value = fmt.Sprintf("<a href='%s%s'>%v</a>", link, mConfig.AdditionalUrlParams, value)
	}

	if dateTimeFieldConfig, dateTimeFieldExists := mConfig.DateTimeFields[field]; dateTimeFieldExists {
		value = sqlutils.ExtractFormattedTime(value, dateTimeFieldConfig)
	}

	if fldCfg.FieldEditor == "bs5switch" {
		_, fieldTag := s.renderFormInputTag(&fldCfg, mConfig, record, value)
		value = fieldTag
	}

	return value, tagAttrs
}
