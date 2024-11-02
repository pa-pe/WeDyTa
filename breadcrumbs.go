package wedyta

func (c *RenderTableImpl) breadcrumbBuilder(config *modelConfig) string {
	breadcrumbStr := `<nav style="--bs-breadcrumb-divider: '>';" aria-label="breadcrumb">` + "\n"
	breadcrumbStr += `  <ol class="breadcrumb">` + "\n"
	breadcrumbStr += `    <li class="breadcrumb-item"><a href="/">Home</a></li>` + "\n"

	if parentModelName, parentExists := config.Parent["modelName"]; parentExists {
		breadcrumbStr += `    <li class="breadcrumb-item"><a href="/render_table/` + parentModelName + `">` + config.ParentConfig.PageTitle + `</a></li>` + "\n"
	}

	breadcrumbStr += `    <li class="breadcrumb-item active" aria-current="page">` + config.PageTitle + ` &nbsp; <i class="bi-arrow-repeat" style="color: grey; cursor: pointer;" onClick="window.location.href = window.location.pathname + window.location.search + window.location.hash;"></i>` + `</li>` + "\n"
	breadcrumbStr += `  </ol>` + "\n"
	breadcrumbStr += `</nav>` + "\n"

	return breadcrumbStr
}
