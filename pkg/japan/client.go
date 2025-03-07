package japan

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-resty/resty/v2"

	"github.com/arisu-archive/memorypack-go"
)

type Client struct {
	client       *resty.Client
	resourcePath string
}

func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) GetResource(ctx context.Context, filePath string) ([]byte, error) {
	resourcePath, err := c.getResourcePath(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource path: %w", err)
	}

	fullPath := resourcePath + "/" + filePath
	resp, err := c.client.R().SetContext(ctx).Get(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download resource: %w", err)
	}
	return resp.Body(), nil
}

func (c *Client) GetResources(ctx context.Context, filter string) ([]string, error) {
	// Getting the TableCatalog.bytes, MediaCatalog.bytes, and so on.
	// Deserialize the catalog file using memorypack-go.
	// Then, apply the filter to the catalog file.
	// Finally, download the resources.
	tableCatalogData, err := c.GetResource(ctx, "TableBundles/TableCatalog.bytes")
	if err != nil {
		return nil, fmt.Errorf("failed to get table catalog: %w", err)
	}

	mediaCatalogData, err := c.GetResource(ctx, "MediaResources/Catalog/MediaCatalog.bytes")
	if err != nil {
		return nil, fmt.Errorf("failed to get media catalog: %w", err)
	}

	bundleDownloadInfoData, err := c.GetResource(ctx, "Android/bundleDownloadInfo.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle download info: %w", err)
	}

	var tableCatalog TableCatalog
	err = memorypack.Deserialize(tableCatalogData, &tableCatalog)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize table catalog: %w", err)
	}

	var mediaCatalog MediaCatalog
	err = memorypack.Deserialize(mediaCatalogData, &mediaCatalog)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize media catalog: %w", err)
	}

	var bundleDownloadInfo BundleDownloadInfo
	err = json.Unmarshal(bundleDownloadInfoData, &bundleDownloadInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle download info: %w", err)
	}

	resources := make([]string, 0)
	for _, tableBundle := range tableCatalog.TableBundles {
		resources = append(resources, "TableBundles/"+tableBundle.Path)
	}
	for _, mediaBundle := range mediaCatalog.MediaBundles {
		resources = append(resources, "MediaResources/"+mediaBundle.Path)
	}
	for _, resource := range bundleDownloadInfo.Files {
		resources = append(resources, "Android/"+resource.Name)
	}

	filteredResources := make([]string, 0)
	for _, resource := range resources {
		matches, matchErr := doublestar.Match(filter, resource)
		if matchErr != nil {
			return nil, fmt.Errorf("failed to match resource: %w", matchErr)
		}
		if matches {
			filteredResources = append(filteredResources, resource)
		}
	}

	return filteredResources, nil
}

func (c *Client) GetVersion(ctx context.Context) (string, error) {
	resp, err := c.client.R().SetContext(ctx).Get(GetClientVersionURL)
	if err != nil || resp.IsError() {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	return string(resp.Body()), nil
}

func (c *Client) getAssetsURL(version string) (string, error) {
	resp, err := c.client.R().Get(fmt.Sprintf(GetAssetsVersionURL, version))
	if err != nil || resp.IsError() {
		return "", fmt.Errorf("failed to send request to get assets version: %w", err)
	}

	var result ServerInfoData
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return result.ServerInfoDataURL, nil
}

func (c *Client) versionCheck(version string) (*VersionCheckResponse, error) {
	assetsURL, err := c.getAssetsURL(version)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets version: %w", err)
	}
	resp, err := c.client.R().Get(assetsURL)
	if err != nil || resp.IsError() {
		return nil, fmt.Errorf("failed to perform version check: %v", resp.Body())
	}

	var result VersionCheckResponse
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &result, nil
}

func (c *Client) getResourcePath(ctx context.Context) (string, error) {
	if c.resourcePath != "" {
		return c.resourcePath, nil
	}

	version, err := c.GetVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	resp, err := c.versionCheck(version)
	if err != nil {
		return "", err
	}

	// TODO: Support incremental update.
	c.resourcePath = resp.ConnectionGroups[0].OverrideConnectionGroups[1].AddressablesCatalogURLRoot
	return c.resourcePath, nil
}
