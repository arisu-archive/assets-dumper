package japan

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/cespare/xxhash/v2"
	"github.com/go-resty/resty/v2"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

var _ resourceapi.Client = (*Client)(nil)

type Client struct {
	client       *resty.Client
	resourcePath string
	retriever    *CatalogRetriever
}

func NewClient(client *resty.Client) *Client {
	c := &Client{
		client: client,
	}
	c.retriever = NewCatalogRetriever(c)
	return c
}

func NewClientWithRetriever(client *resty.Client, retriever *CatalogRetriever) *Client {
	return &Client{
		client:    client,
		retriever: retriever,
	}
}

func (c *Client) DownloadResource(ctx context.Context, filePath string) (io.ReadCloser, int64, error) {
	resourcePath, err := c.getResourcePath(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get resource path: %w", err)
	}

	fullPath := resourcePath + "/" + filePath
	resp, err := c.client.R().
		SetDoNotParseResponse(true).
		SetContext(ctx).
		Get(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download resource: %w", err)
	}
	return resp.RawBody(), resp.RawResponse.ContentLength, nil
}

func (c *Client) getResource(ctx context.Context, filePath string) ([]byte, error) {
	reader, _, err := c.DownloadResource(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource: %w", err)
	}
	return data, nil
}

func (c *Client) ListResources(ctx context.Context, filter string) ([]resourceapi.Resource, error) {
	// First, fetch and parse all catalog files
	resources, err := c.retriever.CollectAllResources(ctx)
	if err != nil {
		return nil, err
	}

	// Then apply the filter
	if filter == "" || filter == "*" {
		return resources, nil
	}

	filteredResources := make([]resourceapi.Resource, 0)
	for _, resource := range resources {
		matches, matchErr := doublestar.Match(filter, resource.Path)
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

func (c *Client) IsResourceCached(_ context.Context, resource resourceapi.Resource, fullPath string) bool {
	// 1. If file not found, download it.
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return true
	}
	// 2. If file found, compare the file hash.
	ourHash, err := c.ComputeHash(fullPath)
	if err != nil {
		return false
	}
	if ourHash != resource.Hash {
		return true
	}
	return false
}

func (*Client) ComputeHash(fullPath string) (string, error) {
	// Read the file, using xxhash.
	reader, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer reader.Close()
	hash := xxhash.New()
	if _, copyErr := io.Copy(hash, reader); copyErr != nil {
		return "", fmt.Errorf("failed to copy file: %w", copyErr)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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
