package service

import (
	"fmt"
	"github.com/pa-pe/wedyta/model"
	"html"
	"strings"
)

func (s *Service) renderFormInputTag(fldCfg *model.FieldParams, record map[string]interface{}, value interface{}) (string, string) {
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

	switch fldCfg.FieldEditor {
	case "textarea":
		htmlTag.WriteString(fmt.Sprintf("<textarea class=\"form-control\" id=\"%s\" name=\"%s\"%s>%v</textarea>", field, field, requiredAttr, value))
	case "input":
		htmlTag.WriteString(fmt.Sprintf("<input class=\"form-control\" type=\"text\" id=\"%s\" name=\"%s\" value=\"%v\"%s>", field, field, value, requiredAttr))
	case "select":
		value_ := takeFieldValueFromRecord(field, record)
		htmlSelect, err := s.RenderRelatedDataSelect(fldCfg.RelatedData, value_, fldCfg.IsRequired)
		if err != nil {
			htmlTag.WriteString("oops")
		} else {
			htmlTag.WriteString(htmlSelect)
		}
	case "summernote":
		htmlTag.WriteString(fmt.Sprintf("<textarea class=\"form-control\" id=\"%s\" name=\"%s\"%s>%v</textarea>", field, field, requiredAttr, value))
	default:
		htmlTag.WriteString("oops, something went wrong")
	}

	return labelTag, htmlTag.String()
}

func (s *Service) RenderRelatedDataSelect(rdCfg *model.RelatedDataConfig, selected interface{}, required bool) (string, error) {
	var records []map[string]interface{}

	if err := s.DB.
		Table(rdCfg.TableName).
		Select([]string{rdCfg.PrimaryKeyFieldName, rdCfg.FieldName}).
		Find(&records).Error; err != nil {
		return "", err
	}

	var htmlSelect strings.Builder
	requiredAttr := ""
	if required {
		requiredAttr = " required"
	}

	htmlSelect.WriteString(`<select class="form-select" name="` + rdCfg.PrimaryKeyFieldName + `"` + requiredAttr + `>` + "\n")
	htmlSelect.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`+"\n", "", "", ""))

	for _, record := range records {
		val := fmt.Sprint(record[rdCfg.PrimaryKeyFieldName])
		text := fmt.Sprint(record[rdCfg.FieldName])

		selectedAttr := ""
		if selected != nil && fmt.Sprint(selected) == val {
			selectedAttr = ` selected`
		}

		htmlSelect.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`+"\n", html.EscapeString(val), selectedAttr, html.EscapeString(text)))
	}

	htmlSelect.WriteString(`</select>` + "\n")
	return htmlSelect.String(), nil
}
