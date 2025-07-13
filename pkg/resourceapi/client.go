package resourceapi

import (
	"context"
	"io"
)

type Client interface {
	DownloadResource(ctx context.Context, resourcePath string) (io.ReadCloser, int64, error)
	DownloadApplication(ctx context.Context) (io.ReadCloser, int64, error)
	DownloadPatch(ctx context.Context, patchPath string) (io.ReadCloser, int64, error)
	ListPatches(ctx context.Context, filter string) ([]Resource, error)
	ListResources(ctx context.Context, filter string) ([]Resource, error)
	GetCatalog(ctx context.Context, catalogType CatalogType) (any, error)
	GetLatestVersion(ctx context.Context) (string, error)
	GetLatestPatchVersion(ctx context.Context) (string, error)
	IsResourceCached(ctx context.Context, resource Resource, fullPath string) bool
	WithVersion(version string) Client
	WithPatchVersion(patchVersion int64) Client
}
