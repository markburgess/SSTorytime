package server

import (
	"embed"
	"io/fs"
)

//go:embed all:public
var content embed.FS

// PublicFS returns the embedded static assets (html/js/icons).
// Theme CSS is served from internal/css via sync.Once, not from here.
func PublicFS() (fs.FS, error) {
	return fs.Sub(content, "public")
}
