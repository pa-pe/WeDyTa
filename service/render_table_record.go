package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (s *Service) RenderTableRecord(ctx *gin.Context) {
	modelName := ctx.Param("modelName")
	recIDstr := ctx.Param("recID")
	recID, err := strconv.ParseInt(recIDstr, 10, 64)
	if err != nil {
		s.SomethingWentWrong(ctx, "Can't ParseInt recID="+recIDstr)
	}

	action := ctx.Param("action")
	isUpdateMode := false
	if action == "" {
		action = "read"
	} else if action == "update" {
		isUpdateMode = true
	} else {
		s.SomethingWentWrong(ctx, "Can't ParseInt action="+action)
	}

	if s.Config.AccessCheckFunc(ctx, modelName, "", action) != true {
		ctx.String(http.StatusForbidden, "Access Denied")
		return
	}

	mConfig := s.loadModelConfig(ctx, modelName, nil)
	if mConfig == nil {
		return
	}

	htmlTable, err := s.RenderModelTableRecord(ctx, mConfig, recID, isUpdateMode)
	if err != nil {
		s.SomethingWentWrong(ctx, fmt.Sprintf("RenderModelTableRecord error: %v", err))
		return
	}

	s.RenderPage(ctx, mConfig, htmlTable)
}

func (s *Service) RenderModelTableRecord(ctx *gin.Context, mConfig *model.ConfigOfModel, recID int64, isUpdateMode bool) (string, error) {
	if mConfig == nil {
		log.Fatalf("Wedyta: RenderModelTableRecord(): mConfig == nil")
	}

	if mConfig.DbTablePrimaryKey == "" {
		err := fmt.Errorf("empty mConfig.DbTablePrimaryKey for model: %s", mConfig.ModelName)
		return "", err
	}

	var record map[string]interface{}
	if err := s.DB.
		Model(&record).
		Table(mConfig.DbTable).
		Where(fmt.Sprintf("%s = %d", mConfig.DbTablePrimaryKey, recID)).
		Where(mConfig.SqlWhere).
		Take(&record).Error; err != nil {
		return "", err
	}

	var htmlTable strings.Builder

	if len(mConfig.EditableFields) > 0 {
		htmlTable.WriteString(`
<script src="https://code.jquery.com/jquery-3.7.1.min.js"></script>
<script src="/wedyta/static/js/wedyta_update.js"></script>
` + mConfig.AdditionalScripts)
	}

	htmlTable.WriteString(`
<style>
table { width: auto !important; }
th { white-space: nowrap; width: 55px; }
td { width: auto !important; }
.form-label { font-weight: bold; }
</style>
`)

	htmlTable.WriteString("<div class=\"col\">\n")

	htmlTable.WriteString(`<` + s.Config.HeadersTag + `>` + mConfig.PageTitle + `</` + s.Config.HeadersTag + `>` + "\n")
	htmlTable.WriteString(s.breadcrumbBuilder(mConfig, fmt.Sprintf("%d", recID)))

	var pkValue string
	value, exists := record[mConfig.DbTablePrimaryKey]
	if exists {
		pkValue = fmt.Sprintf("%v", value)
	}
	if isUpdateMode {
		if pkValue == "" {
			s.SomethingWentWrong(ctx, "Can't take primary key value")
		}

		htmlTable.WriteString("<form id=\"editForm\">\n")
		htmlTable.WriteString(" <input type=\"hidden\" name=\"modelName\" value=\"" + mConfig.ModelName + "\">\n")
		htmlTable.WriteString("<input type=\"hidden\" name=\"id\" value=\"" + pkValue + "\">\n")
	}

	htmlTable.WriteString("<table class='table table-striped mt-3' model='" + mConfig.ModelName + "' record_id='" + pkValue + "'>\n<tbody>\n<tr>\n")

	for _, field := range mConfig.Fields {
		fldCfg := mConfig.FieldConfig[field]

		if isUpdateMode {
			// skip stdRecordControls in isUpdateMode
			if mConfig.ColumnDataFunc[field] == "stdRecordControls" {
				continue
			}

			if !fldCfg.PermitDisplayInUpdateMode {
				continue
			}
		} else {
			if !fldCfg.PermitDisplayInRecordMode {
				continue
			}
		}

		header := fldCfg.Header

		titleStr := ""
		if fldCfg.Title != "" {
			titleStr = fmt.Sprintf(" title='%s'", fldCfg.Title)
		}

		var cache model.RenderTableCache
		cache.RelatedData = make(map[string]string)

		value, tagAttrs := s.renderRecordValue(mConfig, field, record, &cache)
		if isUpdateMode && fldCfg.IsEditable {
			labelTag, fieldTag := s.renderFormInputTag(&fldCfg, record, value)
			htmlTable.WriteString("<tr>\n <td" + tagAttrs + " colspan=\"2\">\n")
			htmlTable.WriteString(labelTag + "<br>\n")
			htmlTable.WriteString(fieldTag + "\n")
			htmlTable.WriteString("</td>\n</tr>\n")
		} else {
			htmlTable.WriteString(fmt.Sprintf("<tr>\n <th%s id=\"header_of_%s\">%s</th>\n <td%s>%v</td>\n</tr>\n", titleStr, field, header, tagAttrs, value))
		}
	}
	htmlTable.WriteString("</tbody>\n</table>\n")

	if isUpdateMode {
		htmlTable.WriteString("<button type=\"button\" class=\"btn btn-primary\" id=\"saveButton\">Save</button>\n")
		htmlTable.WriteString("</form>\n")
	}

	htmlTable.WriteString("</div>\n")

	return htmlTable.String(), nil
}
