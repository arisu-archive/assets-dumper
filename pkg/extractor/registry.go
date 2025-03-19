package extractor

import (
	"context"
	"sync"
)

//nolint:gochecknoglobals // This is a registry
var registry sync.Map

type FileFormat string

const (
	fileFormatZip    FileFormat = ".zip"
	fileFormatSqlite FileFormat = ".db"
)

//nolint:gochecknoinits // This is a registry
func init() {
	registry.Store(fileFormatZip, Extractor(zipExtractor))
	registry.Store(fileFormatSqlite, Extractor(sqliteExtractor))
}

type Extractor func(ctx context.Context, inputPath string) (*Result, error)

func RegisterExtractor(extension FileFormat, extractor Extractor) {
	if _, dup := registry.LoadOrStore(extension, extractor); dup {
		panic("extractor already registered")
	}
}

// GetExtractor returns the appropriate extractor for the given extension.
func extractor(extension FileFormat) Extractor {
	extractor, ok := registry.Load(extension)
	if !ok {
		return nil
	}
	//revive:disable:unchecked-type-assertion // Guaranteed to be an Extractor
	return extractor.(Extractor) //nolint:errcheck // Guaranteed to be an Extractor
}
