package decryption

import (
	"github.com/arisu-archive/plana-flatbuffers/go/flatdata"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
)

type japanFlatDataProvider struct{}

func (*japanFlatDataProvider) GetFlatDataByName(name string) fbsutils.FlatData {
	return flatdata.GetFlatDataByName(name)
}
