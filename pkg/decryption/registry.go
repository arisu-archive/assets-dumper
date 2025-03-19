package decryption

import (
	"context"
	"io"
	"sync"
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
	registry.Store(fileFormatFlatdata, Reader(flatdataReader))
}

type Reader func(ctx context.Context, name string, size int, r io.Reader) (io.Reader, error)

func RegisterDecryptionReader(extension FileFormat, decryptionReader Reader) {
	if _, dup := registry.LoadOrStore(extension, decryptionReader); dup {
		panic("decryption reader already registered")
	}
}

// decryptionReader returns the appropriate decryption reader for the given extension.
func decryptionReader(extension FileFormat) Reader {
	dr, ok := registry.Load(extension)
	if !ok {
		return nil
	}
	//revive:disable:unchecked-type-assertion // Guaranteed to be an decryption reader
	return dr.(Reader) //nolint:errcheck // Guaranteed to be an decryption reader
}
