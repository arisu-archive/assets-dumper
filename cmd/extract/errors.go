package extract

import "errors"

var (
	ErrInvalidInputPath  = errors.New("invalid input path")
	ErrInvalidOutputPath = errors.New("invalid output path")
	ErrInvalidServer     = errors.New("invalid server")
	ErrInvalidKey        = errors.New("invalid key: must be a base64-encoded string")
)
