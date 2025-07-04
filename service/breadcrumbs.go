package service

import "github.com/pa-pe/wedyta/model"

func (s *Service) breadcrumbBuilder(config *model.ConfigOfModel, recID string) string {
	breadcrumbStr := `<nav style="--bs-breadcrumb-divider: '` + s.Config.BreadcrumbsDivider + `';" aria-label="breadcrumb">` + "\n"
	breadcrumbStr += `  <ol class="breadcrumb">` + "\n"
	breadcrumbStr += `    <li class="breadcrumb-item"><a href="` + s.Config.BreadcrumbsRootUrl + `">` + s.Config.BreadcrumbsRootName + `</a></li>` + "\n"

	if parentModelName, parentExists := config.Parent["modelName"]; parentExists {
		breadcrumbStr += `    <li class="breadcrumb-item"><a href="/wedyta/` + parentModelName + `">` + config.ParentConfig.PageTitle + `</a></li>` + "\n"
	}

	breadcrumbStr += `    <li class="breadcrumb-item active" aria-current="page"><a href="/wedyta/` + config.ModelName + `">` + config.PageTitle + `</a>`
	if recID != "" {
		breadcrumbStr += `</li>` + "\n" + `    <li class="breadcrumb-item active" aria-current="page"> #` + recID
	}
	breadcrumbStr += ` &nbsp; <i class="bi-arrow-repeat" style="color: grey; cursor: pointer;" onClick="window.location.href = window.location.pathname + window.location.search + window.location.hash;"></i>` + `</li>` + "\n"
	breadcrumbStr += `  </ol>` + "\n"
	breadcrumbStr += `</nav>` + "\n"

	return breadcrumbStr
}
