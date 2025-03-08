package japan

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arisu-archive/memorypack-go"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

var _ resourceapi.Retriever = (*CatalogRetriever)(nil)

// CatalogRetriever is a client for the catalog.
type CatalogRetriever struct {
	client *Client
}

// NewCatalogRetriever creates a new CatalogRetriever.
func NewCatalogRetriever(client *Client) *CatalogRetriever {
	return &CatalogRetriever{
		client: client,
	}
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

// GetTableResources retrieves and parses table resources from the catalog.
func (cr *CatalogRetriever) GetTableResources(ctx context.Context) ([]resourceapi.Resource, error) {
	tableCatalogData, dlErr := cr.client.getResource(ctx, "TableBundles/TableCatalog.bytes")
	if dlErr != nil {
		return nil, fmt.Errorf("failed to get table catalog: %w", dlErr)
	}

	var tableCatalog TableCatalog
	if err := memorypack.Deserialize(tableCatalogData, &tableCatalog); err != nil {
		return nil, fmt.Errorf("failed to deserialize table catalog: %w", err)
	}

	resources := []resourceapi.Resource{}
	for _, tableBundle := range tableCatalog.TableBundles {
		resources = append(resources, resourceapi.Resource{
			Path: tableBundle.Path,
			Size: tableBundle.Bytes,
			Hash: fmt.Sprintf("%x", tableBundle.Crc),
		})
	}
	return resources, nil
}

// GetMediaResources retrieves and parses media resources from the catalog.
func (cr *CatalogRetriever) GetMediaResources(ctx context.Context) ([]resourceapi.Resource, error) {
	mediaCatalogData, dlErr := cr.client.getResource(ctx, "MediaResources/Catalog/MediaCatalog.bytes")
	if dlErr != nil {
		return nil, fmt.Errorf("failed to get media catalog: %w", dlErr)
	}

	var mediaCatalog MediaCatalog
	if err := memorypack.Deserialize(mediaCatalogData, &mediaCatalog); err != nil {
		return nil, fmt.Errorf("failed to deserialize media catalog: %w", err)
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
	bundleDownloadInfoData, dlErr := cr.client.getResource(ctx, "Android/bundleDownloadInfo.json")
	if dlErr != nil {
		return nil, fmt.Errorf("failed to get bundle download info: %w", dlErr)
	}

	var bundleDownloadInfo BundleDownloadInfo
	if err := json.Unmarshal(bundleDownloadInfoData, &bundleDownloadInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle download info: %w", err)
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
