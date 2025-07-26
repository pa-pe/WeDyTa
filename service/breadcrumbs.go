package service

import (
	"fmt"
	"github.com/pa-pe/wedyta/model"
	"log"
)

func (s *Service) breadcrumbBuilder(mConfig *model.ConfigOfModel, recID string, action string) string {
	breadcrumbStr := `<nav style="--bs-breadcrumb-divider: '` + s.Config.BreadcrumbsDivider + `';" aria-label="breadcrumb">` + "\n"
	breadcrumbStr += `  <ol class="breadcrumb">` + "\n"
	breadcrumbStr += `    <li class="breadcrumb-item"><a href="` + s.Config.BreadcrumbsRootUrl + `">` + s.Config.BreadcrumbsRootName + `</a></li>` + "\n"

	if mConfig.HasParent {
		breadcrumbStr += s.renderParentBreadcrumb(mConfig)
	}

	breadcrumbStr += `    <li class="breadcrumb-item active" aria-current="page"><a href="/wedyta/` + mConfig.ModelName + `">` + mConfig.PageTitle + `</a>`

	if recID != "" {
		breadcrumbStr += `</li>` + "\n" + `    <li class="breadcrumb-item active" aria-current="page"> #` + recID
	}
	switch action {
	case "read records":
	case "read record":
	case "create":
		breadcrumbStr += `</li>` + "\n" + `    <li class="breadcrumb-item active" aria-current="page"> ` + "create record"
	case "update":
		breadcrumbStr += `</li>` + "\n" + `    <li class="breadcrumb-item active" aria-current="page"> ` + "update record"
	}

	breadcrumbStr += ` &nbsp; <i class="bi-arrow-repeat" style="color: grey; cursor: pointer;" onClick="window.location.href = window.location.pathname + window.location.search + window.location.hash;" title="Refresh page"></i>` + `</li>` + "\n"
	breadcrumbStr += `  </ol>` + "\n"
	breadcrumbStr += `</nav>` + "\n"

	return breadcrumbStr
}

func (s *Service) renderParentBreadcrumb(mConfig *model.ConfigOfModel) string {
	breadcrumbStr := ""

	parentMC := mConfig.ParentConfig
	breadcrumbStr += `    <li class="breadcrumb-item"><a href="/wedyta/` + parentMC.ModelName + `">` + mConfig.ParentConfig.PageTitle + `</a></li>` + "\n"
	if mConfig.Parent["queryVariableName"] != "" && mConfig.Parent["queryVariableValue"] != "" {
		value := ""
		if mConfig.ParentConfig.Breadcrumb.LabelField != "" {
			var err error
			value, err = s.takeLabelFieldValue(parentMC.DbTable, parentMC.DbTablePrimaryKey, mConfig.Parent["queryVariableValue"], mConfig.ParentConfig.Breadcrumb.LabelField)
			if err != nil {
				log.Printf("Error taking label field: %s", err.Error())
			}
		} else {
			value = "#" + mConfig.Parent["queryVariableValue"]
		}
		breadcrumbStr += `    <li class="breadcrumb-item"><a href="/wedyta/` + parentMC.ModelName + `/` + mConfig.Parent["queryVariableValue"] + parentMC.AdditionalUrlParams + `">` + value + `</a></li>` + "\n"
	}

	if parentMC.HasParent {
		breadcrumbStr = s.renderParentBreadcrumb(parentMC) + breadcrumbStr
	}

	return breadcrumbStr
}

func (s *Service) takeLabelFieldValue(table, pkField, pkValue, labelField string) (string, error) {
	var record map[string]interface{}
	if err := s.DB.
		Model(&record).
		Table(table).
		Where(fmt.Sprintf("%s = %s", pkField, pkValue)).
		Take(&record).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", record[labelField]), nil
}
