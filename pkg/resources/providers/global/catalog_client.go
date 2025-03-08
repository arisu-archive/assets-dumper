package global

import (
	"context"
	"encoding/json"
	"fmt"
)

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

func (cr *CatalogRetriever) getResource(ctx context.Context, path string) ([]byte, error) {
	return cr.client.DownloadResourceToFile(ctx, path)
}

// GetCatalog retrieves and parses catalog resources from the catalog.
func (cr *CatalogRetriever) GetTableCatalog(ctx context.Context) (*TableCatalog, error) {
	tableCatalogData, err := cr.getResource(ctx, "Catalog/TableBundles/TableCatalog.bytes")
	if err != nil {
		return nil, fmt.Errorf("failed to get table catalog: %w", err)
	}

	var tableCatalog TableCatalog
	if decodeErr := json.Unmarshal(tableCatalogData, &tableCatalog); decodeErr != nil {
		return nil, fmt.Errorf("failed to unmarshal table catalog: %w", decodeErr)
	}

	return &tableCatalog, nil
}

func (cr *CatalogRetriever) GetMediaCatalog(ctx context.Context) (*MediaCatalog, error) {
	mediaCatalogData, err := cr.getResource(ctx, "Catalog/MediaResources/MediaCatalog.bytes")
	if err != nil {
		return nil, fmt.Errorf("failed to get media catalog: %w", err)
	}

	var mediaCatalog MediaCatalog
	if decodeErr := json.Unmarshal(mediaCatalogData, &mediaCatalog); decodeErr != nil {
		return nil, fmt.Errorf("failed to unmarshal media catalog: %w", decodeErr)
	}

	return &mediaCatalog, nil
}

func (cr *CatalogRetriever) GetBundleDownloadInfo(ctx context.Context) (any, error) {
	bundleDownloadInfoData, err := cr.getResource(ctx, "Catalog/Android/catalog_Android.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle download info: %w", err)
	}

	var bundleDownloadInfo any
	if decodeErr := json.Unmarshal(bundleDownloadInfoData, &bundleDownloadInfo); decodeErr != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle download info: %w", decodeErr)
	}

	return &bundleDownloadInfo, nil
}
