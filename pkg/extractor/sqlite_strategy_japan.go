package extractor

import (
	"context"

	"github.com/arisu-archive/plana-flatbuffers/go/excel"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
)

// JapanExcelProvider implements the ExcelProvider interface for Japan version.
type JapanExcelProvider struct{}

// GetExcelByName returns the excel flatdata for the given name (Japan version).
func (*JapanExcelProvider) GetExcelByName(name string) fbsutils.FlatData {
	return excel.GetFlatDataByName(name)
}

// Japan excel provider singleton.
var japanExcelProvider = &JapanExcelProvider{} //nolint:gochecknoglobals // internal singleton

func sqliteExtractorJapan(ctx context.Context, inputPath string) (*Result, error) {
	return sqliteExtractorCommon(ctx, japanExcelProvider, inputPath)
}
