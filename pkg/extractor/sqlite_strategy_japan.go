package extractor

import (
	"context"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
	"github.com/arisu-archive/plana-flatbuffers/go/excel"
)

// JapanExcelProvider implements the ExcelProvider interface for Japan version
type JapanExcelProvider struct{}

// GetExcelByName returns the excel flatdata for the given name (Japan version)
func (p *JapanExcelProvider) GetExcelByName(name string) fbsutils.FlatData {
	return excel.GetFlatDataByName(name)
}

// Japan excel provider singleton
var japanExcelProvider = &JapanExcelProvider{}

func sqliteExtractorJapan(ctx context.Context, inputPath string) (*Result, error) {
	return sqliteExtractorCommon(ctx, japanExcelProvider, inputPath)
}
