package extractor

import (
	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
)

// ExcelProvider defines the interface for getting excel flatdata by name
type ExcelProvider interface {
	GetExcelByName(name string) fbsutils.FlatData
}
