package decryption

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

//nolint:gochecknoglobals // This is a registry
var registry sync.Map

type FileFormat string

const (
	// We can safely assumed the .bytes file is an encrypted flatdata file.
	fileFormatFlatdata FileFormat = ".bytes"
)

//nolint:gochecknoinits // This is a registry
func init() {
	registry.Store(
		decryptorKey(resourceapi.ServerGlobal, fileFormatFlatdata), flatdataReader(&globalFlatDataProvider{}),
	)
	registry.Store(
		decryptorKey(resourceapi.ServerJapan, fileFormatFlatdata), flatdataReader(&japanFlatDataProvider{}),
	)
}

func flatdataReader(p FlatDataProvider) Reader {
	return func(ctx context.Context, name string, size uint64, r io.Reader) (io.Reader, error) {
		return flatdataReaderCommon(ctx, flatdataReaderOptions{
			provider: p,
			name:     name,
			size:     size,
			r:        r,
		})
	}
}

func decryptorKey(server resourceapi.Server, extension FileFormat) string {
	return fmt.Sprintf("%s:%s", server, extension)
}

type Reader func(ctx context.Context, name string, size uint64, r io.Reader) (io.Reader, error)

func RegisterDecryptionReader(server resourceapi.Server, extension FileFormat, decryptionReader Reader) {
	if _, dup := registry.LoadOrStore(decryptorKey(server, extension), decryptionReader); dup {
		panic("decryption reader already registered")
	}
}

// decryptionReader returns the appropriate decryption reader for the given extension.
func decryptionReader(server resourceapi.Server, extension FileFormat) Reader {
	dr, ok := registry.Load(decryptorKey(server, extension))
	if !ok {
		return nil
	}
	//revive:disable:unchecked-type-assertion // Guaranteed to be an decryption reader
	return dr.(Reader) //nolint:errcheck // Guaranteed to be an decryption reader
}
