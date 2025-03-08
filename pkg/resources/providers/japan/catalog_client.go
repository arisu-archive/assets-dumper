package japan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/arisu-archive/memorypack-go"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

// CatalogRetriever is a client for the catalog.
type CatalogRetriever struct {
	client resourceapi.Client
}

// NewCatalogRetriever creates a new CatalogRetriever.
func NewCatalogRetriever(client resourceapi.Client) *CatalogRetriever {
	return &CatalogRetriever{
		client: client,
	}
}

func (cr *CatalogRetriever) getResource(ctx context.Context, path string) ([]byte, error) {
	r, _, dlErr := cr.client.DownloadResource(ctx, path)
	if dlErr != nil {
		return nil, fmt.Errorf("failed to get resource: %w", dlErr)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource: %w", err)
	}
	return data, nil
}

// CollectAllResources collects all resources from the catalog.
func (cr *CatalogRetriever) CollectAllResources(ctx context.Context) ([]resourceapi.Resource, error) {
	// Get table catalog resources
	tableResources, err := cr.GetTableResources(ctx)
	if err != nil {
		return nil, err
	}

	// Get media catalog resources
	mediaResources, err := cr.GetMediaResources(ctx)
	if err != nil {
		return nil, err
	}

	// Get bundle download info resources
	bundleResources, err := cr.GetBundleResources(ctx)
	if err != nil {
		return nil, err
	}

	// Combine all resources
	totalSize := len(tableResources) + len(mediaResources) + len(bundleResources)
	resources := make([]resourceapi.Resource, 0, totalSize)
	resources = append(resources, tableResources...)
	resources = append(resources, mediaResources...)
	resources = append(resources, bundleResources...)

	return resources, nil
}

// GetCatalog retrieves and parses catalog resources from the catalog.
func (cr *CatalogRetriever) GetTableCatalog(ctx context.Context) (*TableCatalog, error) {
	tableCatalogData, err := cr.getResource(ctx, "TableBundles/TableCatalog.bytes")
	if err != nil {
		return nil, fmt.Errorf("failed to get table catalog: %w", err)
	}

	var tableCatalog TableCatalog
	if decodeErr := memorypack.Deserialize(tableCatalogData, &tableCatalog); decodeErr != nil {
		return nil, fmt.Errorf("failed to unmarshal table catalog: %w", decodeErr)
	}

	return &tableCatalog, nil
}

func (cr *CatalogRetriever) GetMediaCatalog(ctx context.Context) (*MediaCatalog, error) {
	mediaCatalogData, err := cr.getResource(ctx, "MediaResources/Catalog/MediaCatalog.bytes")
	if err != nil {
		return nil, fmt.Errorf("failed to get media catalog: %w", err)
	}

	var mediaCatalog MediaCatalog
	if decodeErr := memorypack.Deserialize(mediaCatalogData, &mediaCatalog); decodeErr != nil {
		return nil, fmt.Errorf("failed to unmarshal media catalog: %w", decodeErr)
	}

	return &mediaCatalog, nil
}

func (cr *CatalogRetriever) GetBundleDownloadInfo(ctx context.Context) (*BundleDownloadInfo, error) {
	bundleDownloadInfoData, err := cr.getResource(ctx, "Android/bundleDownloadInfo.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle download info: %w", err)
	}

	var bundleDownloadInfo BundleDownloadInfo
	if decodeErr := json.Unmarshal(bundleDownloadInfoData, &bundleDownloadInfo); decodeErr != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle download info: %w", decodeErr)
	}

	return &bundleDownloadInfo, nil
}

// GetTableResources retrieves and parses table resources from the catalog.
func (cr *CatalogRetriever) GetTableResources(ctx context.Context) ([]resourceapi.Resource, error) {
	tableCatalog, err := cr.GetTableCatalog(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get table catalog: %w", err)
	}

	resources := []resourceapi.Resource{}
	for _, tableBundle := range tableCatalog.TableBundles {
		resources = append(resources, resourceapi.Resource{
			Path: "TableBundles/" + tableBundle.Path,
			Size: tableBundle.Bytes,
			Hash: fmt.Sprintf("%x", tableBundle.Crc),
		})
	}
	return resources, nil
}

// GetMediaResources retrieves and parses media resources from the catalog.
func (cr *CatalogRetriever) GetMediaResources(ctx context.Context) ([]resourceapi.Resource, error) {
	mediaCatalog, err := cr.GetMediaCatalog(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get media catalog: %w", err)
	}

	resources := []resourceapi.Resource{}
	for _, mediaBundle := range mediaCatalog.MediaBundles {
		resources = append(resources, resourceapi.Resource{
			Path: "MediaResources/" + mediaBundle.Path,
			Size: mediaBundle.Bytes,
			Hash: fmt.Sprintf("%x", mediaBundle.Crc),
		})
	}
	return resources, nil
}

// GetBundleResources retrieves and parses bundle resources from the catalog.
func (cr *CatalogRetriever) GetBundleResources(ctx context.Context) ([]resourceapi.Resource, error) {
	bundleDownloadInfo, err := cr.GetBundleDownloadInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle download info: %w", err)
	}

	resources := []resourceapi.Resource{}
	for _, resource := range bundleDownloadInfo.Files {
		resources = append(resources, resourceapi.Resource{
			Path: "Android/" + resource.Name,
			Size: resource.Size,
			Hash: fmt.Sprintf("%x", resource.Crc),
		})
	}
	return resources, nil
}
