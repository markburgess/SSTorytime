package demopocs

import "errors"

var (
	ErrGetNodesMismatch   = errors.New("demo_pocs/dotest_getnodes: lookup mismatch")
	ErrEntireConeNoEnds   = errors.New("demo_pocs/dotest_entirecone: missing end points")
	ErrEntireConeLeftSz   = errors.New("demo_pocs/dotest_entirecone: left size mismatch")
	ErrEntireConeLeftSet  = errors.New("demo_pocs/dotest_entirecone: left set mismatch")
	ErrEntireConeRightSz  = errors.New("demo_pocs/dotest_entirecone: right size mismatch")
	ErrEntireConeRightSet = errors.New("demo_pocs/dotest_entirecone: right set mismatch")
)
