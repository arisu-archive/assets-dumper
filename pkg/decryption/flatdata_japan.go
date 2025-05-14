package decryption

import (
	"context"
	"io"

	fbsutils "github.com/arisu-archive/bluearchive-fbs-utils"
	"github.com/arisu-archive/plana-flatbuffers/go/flatdata"
)

type japanFlatDataProvider struct{}

func (p *japanFlatDataProvider) GetFlatDataByName(name string) fbsutils.FlatData {
	return flatdata.GetFlatDataByName(name)
}

func flatdataReaderJapan(ctx context.Context, name string, size uint64, r io.Reader) (io.Reader, error) {
	return flatdataReaderCommon(ctx, &japanFlatDataProvider{}, name, size, r)
}
