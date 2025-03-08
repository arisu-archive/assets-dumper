package resourceapi

import (
	"context"
	"io"
)

type Client interface {
	DownloadResource(ctx context.Context, resourcePath string) (io.ReadCloser, int64, error)
	ListResources(ctx context.Context, filter string) ([]Resource, error)
	GetVersion(ctx context.Context) (string, error)
}
