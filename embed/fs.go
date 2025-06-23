package embed

import "embed"

//go:embed static/* templates/default.tmpl
var EmbeddedFiles embed.FS
