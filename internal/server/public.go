package server

import (
	"embed"
	"io/fs"
)

//go:embed all:public
var content embed.FS

// PublicFS returns embedded static assets (html/js/css/icons).
// Theme CSS is produced by go:generate (go run ./internal/libexec/cssgen) into public/.
func PublicFS() (fs.FS, error) {
	return fs.Sub(content, "public")
}
