package main

import "errors"

var (
	ErrQueryRequired   = errors.New("search: query required")
	ErrChapterRequired = errors.New("remove: chapter name required")
	ErrPathsolveArgs   = errors.New("pathsolve: --begin and --end required")
	ErrPositiveN       = errors.New("n must be a positive integer")
	ErrMigrateVersion  = errors.New("migrate version")
	ErrExamplesLoad    = errors.New("examples load")
	ErrExamplesRun     = errors.New("examples run")
	ErrGraphReport     = errors.New("graph-report")
	ErrNotes           = errors.New("notes")
	ErrPathsolve       = errors.New("pathsolve")
	ErrRemove          = errors.New("remove")
	ErrSearch          = errors.New("search")
	ErrText2N4L        = errors.New("text2n4l")
)
