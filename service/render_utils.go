package service

import "strings"

func (s *Service) wrapBsAccordion(content, idPrefix, header string) string {
	var formBuilder strings.Builder

	formBuilder.WriteString(`
	<div class="accordion" id="` + idPrefix + `Accordion">
        <div class="accordion-item">
            <` + s.Config.HeadersTag + ` class="accordion-header" id="` + idPrefix + `Heading">
                <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#` + idPrefix + `Collapse" aria-expanded="false" aria-controls="` + idPrefix + `Collapse">
                    <i class="bi-plus-square"></i> &nbsp; ` + header + `
                </button>
            </` + s.Config.HeadersTag + `>
            <div id="` + idPrefix + `Collapse" class="accordion-collapse collapse" aria-labelledby="` + idPrefix + `Heading" data-bs-parent="#` + idPrefix + `Accordion">
                <div class="accordion-body" style="background: rgba(128,128,128,0.1);">
`)

	formBuilder.WriteString(content)
	formBuilder.WriteString("</div>\n</div>\n</div>\n")

	return formBuilder.String()
}
