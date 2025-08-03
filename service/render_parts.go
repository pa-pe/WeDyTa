package service

import (
	"fmt"
	"github.com/pa-pe/wedyta/model"
	"html"
	"strings"
)

func (s *Service) renderFormInputTag(fldCfg *model.FieldParams, mConfig *model.ConfigOfModel, record map[string]interface{}, value interface{}) (string, string) {
	field := fldCfg.Field
	var htmlTag strings.Builder

	titleStr := ""
	if fldCfg.Title != "" {
		titleStr = fmt.Sprintf(" title='%s'", fldCfg.Title)
	}

	requiredAttr := ""
	requiredLabel := ""
	if fldCfg.IsRequired {
		requiredAttr = " required"
		requiredLabel = ` <span class="required-label">(required)</span>`
	}

	labelTag := fmt.Sprintf("<label%s for=\"%s\" class=\"form-label\" id=\"header_of_%s\">%s</label>%s", titleStr, field, field, fldCfg.Header, requiredLabel)

	var value_ interface{}
	if record == nil {
		value_ = value
	} else {
		value_ = takeFieldValueFromRecord(field, record)
	}

	switch fldCfg.FieldEditor {
	case "textarea":
		htmlTag.WriteString(fmt.Sprintf("<textarea class=\"form-control\" id=\"%s\" name=\"%s\"%s>%v</textarea>", field, field, requiredAttr, value))
	case "input":
		htmlTag.WriteString(fmt.Sprintf("<input class=\"form-control\" type=\"text\" id=\"%s\" name=\"%s\" value=\"%v\"%s>", field, field, value, requiredAttr))
	case "select":
		htmlSelect, err := s.RenderRelatedDataSelect(fldCfg.RelatedData, value_, fldCfg.IsRequired)
		if err != nil {
			htmlTag.WriteString("oops")
		} else {
			htmlTag.WriteString(htmlSelect)
		}
	case "summernote":
		htmlTag.WriteString(fmt.Sprintf("<textarea class=\"form-control\" id=\"%s\" name=\"%s\"%s>%v</textarea>", field, field, requiredAttr, value))
	case "bs5switch":
		var pkValue string
		if record != nil {
			pkValueI, exists := record[mConfig.DbTablePrimaryKey]
			if exists {
				pkValue = fmt.Sprintf("%v", pkValueI)
			}
		}

		checked := ""
		if fmt.Sprintf("%v", value_) == "1" {
			checked = " checked"
		}

		disabled := " disabled"
		if fldCfg.IsEditable {
			disabled = ""
		}
		htmlTag.WriteString(fmt.Sprintf("<div class=\"form-check form-switch\"><input class=\"form-check-input\" type=\"checkbox\" role=\"switch\" name=\"%s\" rec_id=\"%s\" id=\"%s_%s\"%s%s></div>", field, pkValue, field, pkValue, checked, disabled))
	default:
		htmlTag.WriteString("oops, something went wrong")
	}

	return labelTag, htmlTag.String()
}

func (s *Service) RenderRelatedDataSelect(rdCfg *model.RelatedDataEntry, selected interface{}, required bool) (string, error) {
	var records []map[string]interface{}

	if rdCfg.RawSql != "" {
		if err := s.DB.
			Raw(rdCfg.RawSql).
			Scan(&records).Error; err != nil {
			return "", err
		}
	} else {
		if err := s.DB.
			Table(rdCfg.Table).
			Select([]string{rdCfg.KeyField, rdCfg.ValueField}).
			Order(rdCfg.OrderBy).
			Find(&records).Error; err != nil {
			return "", err
		}
	}

	var htmlSelect strings.Builder
	requiredAttr := ""
	if required {
		requiredAttr = " required"
	}

	htmlSelect.WriteString(`<select class="form-select" name="` + rdCfg.KeyField + `"` + requiredAttr + `>` + "\n")
	htmlSelect.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`+"\n", "0", "", ""))

	for _, record := range records {
		id := fmt.Sprint(record[rdCfg.KeyField])
		text := fmt.Sprint(record[rdCfg.ValueField])

		selectedAttr := ""
		if selected != nil && fmt.Sprint(selected) == id {
			selectedAttr = ` selected`
		}

		htmlSelect.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`+"\n", html.EscapeString(id), selectedAttr, html.EscapeString(text)))
	}

	htmlSelect.WriteString(`</select>` + "\n")
	return htmlSelect.String(), nil
}
