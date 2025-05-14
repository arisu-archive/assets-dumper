package decryption

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

type Client interface {
	DecryptionReader(ctx context.Context, name string, size uint64, r io.Reader) (io.Reader, error)
}

type client struct {
	server resourceapi.Server
}

func New(server resourceapi.Server) Client {
	return &client{
		server: server,
	}
}

func (c *client) DecryptionReader(ctx context.Context, name string, size uint64, r io.Reader) (io.Reader, error) {
	slog.DebugContext(ctx, "Setting up decryption reader", "name", name)
	decryptor := decryptionReader(c.server, FileFormat(filepath.Ext(name)))
	if decryptor == nil {
		return nil, fmt.Errorf("decryptor not found for file: %s", name)
	}

	return decryptor(ctx, name, size, r)
}
