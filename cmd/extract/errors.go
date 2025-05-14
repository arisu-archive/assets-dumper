package extract

import "errors"

var (
	ErrInvalidInputPath  = errors.New("invalid input path")
	ErrInvalidOutputPath = errors.New("invalid output path")
	ErrInvalidServer     = errors.New("invalid server")
)
