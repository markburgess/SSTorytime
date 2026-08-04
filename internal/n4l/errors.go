package n4l

import "errors"

var (
	ErrNoInputFiles    = errors.New("n4l: at least one input file required")
	ErrChapterConflict = errors.New("chapter conflict; use --force to override")
	ErrNilContext      = errors.New("n4l: nil context")
)
