package decryption

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
)

type Client interface {
	DecryptionReader(ctx context.Context, name string, size int, r io.Reader) (io.Reader, error)
}

type client struct{}

func New() Client {
	return &client{}
}

func (*client) DecryptionReader(ctx context.Context, name string, size int, r io.Reader) (io.Reader, error) {
	slog.DebugContext(ctx, "Setting up decryption reader", "name", name)
	decryptor := decryptionReader(FileFormat(filepath.Ext(name)))
	if decryptor == nil {
		return nil, fmt.Errorf("decryptor not found for file: %s", name)
	}

	return decryptor(ctx, name, size, r)
}
