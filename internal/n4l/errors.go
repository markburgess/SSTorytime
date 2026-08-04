package n4l

import "errors"

// Package-level error table.
var (
	ErrNoInputFiles    = errors.New("n4l: at least one input file required")
	ErrChapterConflict = errors.New("chapter conflict; use --force to override")
)
