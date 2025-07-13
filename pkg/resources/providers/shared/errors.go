package shared

import "errors"

var (
	ErrInvalidBuildVersion  = errors.New("invalid build version")
	ErrPatchVersionNotFound = errors.New("patch version not found")
)
