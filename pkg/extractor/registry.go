package extractor

import (
	"context"
	"fmt"
	"sync"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

//nolint:gochecknoglobals // This is a registry
var registry sync.Map

type FileFormat string

const (
	fileFormatZip    FileFormat = ".zip"
	fileFormatSqlite FileFormat = ".db"
)

func extractorKey(server resourceapi.Server, format FileFormat) string {
	return fmt.Sprintf("%s:%s", server, format)
}

//nolint:gochecknoinits // This is a registry
func init() {
	// Global server extractors
	registry.Store(extractorKey(resourceapi.ServerGlobal, fileFormatZip), Extractor(zipExtractor))
	registry.Store(extractorKey(resourceapi.ServerGlobal, fileFormatSqlite), Extractor(sqliteExtractor))

	// Japan server extractors
	registry.Store(extractorKey(resourceapi.ServerJapan, fileFormatZip), Extractor(zipExtractor))
	registry.Store(extractorKey(resourceapi.ServerJapan, fileFormatSqlite), Extractor(sqliteExtractorJapan))
}

type Extractor func(ctx context.Context, inputPath string) (*Result, error)

func RegisterExtractor(server resourceapi.Server, extension FileFormat, extractor Extractor) {
	if _, dup := registry.LoadOrStore(extractorKey(server, extension), extractor); dup {
		panic("extractor already registered")
	}
}

// GetExtractor returns the appropriate extractor for the given extension and server.
func extractor(server resourceapi.Server, extension FileFormat) Extractor {
	extractor, ok := registry.Load(extractorKey(server, extension))
	if !ok {
		return nil
	}
	//revive:disable:unchecked-type-assertion // Guaranteed to be an Extractor
	return extractor.(Extractor) //nolint:errcheck // Guaranteed to be an Extractor
}
