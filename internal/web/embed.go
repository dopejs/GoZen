package web

import "embed"

//go:embed all:dist
var staticFS embed.FS
