package extractor

import (
	"context"

	"github.com/arisu-archive/arona-flatbuffers/go/excel"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
)

// GlobalExcelProvider implements the ExcelProvider interface for global version.
type GlobalExcelProvider struct{}

// GetExcelByName returns the excel flatdata for the given name (global version).
func (*GlobalExcelProvider) GetExcelByName(name string) fbsutils.FlatData {
	return excel.GetFlatDataByName(name)
}

// Global excel provider singleton.
var globalExcelProvider = &GlobalExcelProvider{} //nolint:gochecknoglobals // internal singleton

func sqliteExtractor(ctx context.Context, inputPath string) (*Result, error) {
	return sqliteExtractorCommon(ctx, globalExcelProvider, inputPath)
}
