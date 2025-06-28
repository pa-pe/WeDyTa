package service

import (
	"fmt"
	"github.com/pa-pe/wedyta/model"
	"html"
	"strings"
)

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
