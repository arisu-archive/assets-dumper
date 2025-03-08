package resourceapi

import (
	"context"
	"io"
)

type Client interface {
	DownloadResource(ctx context.Context, resourcePath string) (io.ReadCloser, int64, error)
	ListResources(ctx context.Context, filter string) ([]Resource, error)
	GetCatalog(ctx context.Context, catalogType CatalogType) (any, error)
	GetVersion(ctx context.Context) (string, error)
	GetPatchVersion(ctx context.Context) (string, error)
	IsResourceCached(ctx context.Context, resource Resource, fullPath string) bool
	WithVersion(version string) Client
}
