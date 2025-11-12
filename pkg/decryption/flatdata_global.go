package decryption

import (
	"context"
	"io"

	"github.com/arisu-archive/arona-flatbuffers/go/flatdata"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
)

type globalFlatDataProvider struct{}

func (*globalFlatDataProvider) GetFlatDataByName(name string) fbsutils.FlatData {
	return flatdata.GetFlatDataByName(name)
}

func flatdataReader(ctx context.Context, name string, size uint64, r io.Reader) (io.Reader, error) {
	return flatdataReaderCommon(ctx, flatdataReaderOptions{
		provider: &globalFlatDataProvider{},
		name:     name,
		size:     size,
		r:        r,
	})
}
