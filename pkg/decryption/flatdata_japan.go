package decryption

import (
	"context"
	"io"

	"github.com/arisu-archive/plana-flatbuffers/go/flatdata"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
)

type japanFlatDataProvider struct{}

func (*japanFlatDataProvider) GetFlatDataByName(name string) fbsutils.FlatData {
	return flatdata.GetFlatDataByName(name)
}

func flatdataReaderJapan(ctx context.Context, name string, size uint64, r io.Reader) (io.Reader, error) {
	return flatdataReaderCommon(ctx, flatdataReaderOptions{
		provider: &japanFlatDataProvider{},
		name:     name,
		size:     size,
		r:        r,
	})
}
