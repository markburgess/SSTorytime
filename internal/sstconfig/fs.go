// Package sstconfig embeds default arrow/annotation config (*.sst next to this
// file). Pass a different fs.FS only when the user sets --config.
package sstconfig

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

//go:embed *.sst
var embedded embed.FS

// DefaultFiles are loaded in order during N4L configuration.
var DefaultFiles = []string{
	"arrows-LT-1.sst",
	"arrows-NR-0.sst",
	"arrows-CN-2.sst",
	"arrows-EP-3.sst",
	"annotations.sst",
	"closures.sst",
}

// BookmarksFile is optional bookmark definitions.
const BookmarksFile = "bookmarks.sst"

var (
	ErrStat   = errors.New("sstconfig")
	ErrNotDir = errors.New("sstconfig: not a directory")
)

// Default returns the embedded default configuration filesystem.
func Default() fs.FS {
	return embedded
}

// Dir returns an fs.FS rooted at an on-disk directory (user override).
func Dir(path string) (fs.FS, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStat, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrNotDir, path)
	}
	return os.DirFS(path), nil
}

// Resolve picks the config FS: explicit dir if non-empty, otherwise embedded defaults.
func Resolve(explicitDir string) (fs.FS, error) {
	if explicitDir == "" {
		return Default(), nil
	}
	return Dir(explicitDir)
}

// ReadFile reads a single config file from fsys.
func ReadFile(fsys fs.FS, name string) ([]byte, error) {
	return fs.ReadFile(fsys, name)
}

// Open opens a named file in fsys.
func Open(fsys fs.FS, name string) (fs.File, error) {
	return fsys.Open(name)
}
