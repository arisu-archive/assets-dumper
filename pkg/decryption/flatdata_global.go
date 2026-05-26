package decryption

import (
	"github.com/arisu-archive/arona-flatbuffers/go/flatdata"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
)

type globalFlatDataProvider struct{}

func (*globalFlatDataProvider) GetFlatDataByName(name string) fbsutils.FlatData {
	return flatdata.GetFlatDataByName(name)
}
