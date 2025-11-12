package japan

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-resty/resty/v2"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

var _ resourceapi.Client = (*Client)(nil)

type Client struct {
	client       *resty.Client
	resourcePath string
	retriever    *CatalogRetriever
	version      string
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

func (c *Client) WithVersion(version string) resourceapi.Client {
	c.version = version
	return c
}

func (c *Client) WithPatchVersion(_ int64) resourceapi.Client {
	return c
}

func (c *Client) GetCatalog(ctx context.Context, catalogType resourceapi.CatalogType) (any, error) {
	switch catalogType {
	case resourceapi.CatalogTypeTableBundle:
		return c.retriever.GetTableCatalog(ctx)
	case resourceapi.CatalogTypeMediaResources:
		return c.retriever.GetMediaCatalog(ctx)
	case resourceapi.CatalogTypeBundleDownloadInfo:
		return c.retriever.GetBundleDownloadInfo(ctx)
	default:
		return nil, fmt.Errorf("unknown catalog type: %s", catalogType)
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

func (c *Client) DownloadApplication(ctx context.Context) (io.ReadCloser, int64, error) {
	version, err := c.GetLatestVersion(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get version: %w", err)
	}
	fullPath := fmt.Sprintf(APKTemplateURL, version)
	slog.DebugContext(ctx, "DownloadApplication", "fullPath", fullPath)
	resp, err := c.client.R().SetDoNotParseResponse(true).SetContext(ctx).Get(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download application: %w", err)
	}
	return resp.RawBody(), resp.RawResponse.ContentLength, nil
}

func (c *Client) ListResources(ctx context.Context, filter string) ([]resourceapi.Resource, error) {
	// First, fetch and parse all catalog files
	resources, err := c.retriever.CollectAllResources(ctx)
	if err != nil {
		return nil, err
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

func (*Client) DownloadPatch(_ context.Context, _ string) (io.ReadCloser, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (*Client) ListPatches(_ context.Context, _ string) ([]resourceapi.Resource, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) GetLatestVersion(ctx context.Context) (string, error) {
	resp, err := c.client.R().SetContext(ctx).Get(GetClientVersionURL)
	if err != nil || resp.IsError() {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	return string(resp.Body()), nil
}

func (c *Client) GetLatestPatchVersion(ctx context.Context) (string, error) {
	// Get the patch version from the AddressablesCatalogURLRoot
	rootURL, err := c.getResourcePath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get resource path: %w", err)
	}

	// Get it from the path. Last part of the path is the patch version.
	parts := strings.Split(rootURL, "/")
	patchVersion := parts[len(parts)-1]
	return patchVersion, nil
}

func (c *Client) IsResourceCached(ctx context.Context, resource resourceapi.Resource, fullPath string) bool {
	// 1. If file not found, download it.
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return false
	}
	// 2. If file found, compare the file hash.
	ourHash, err := c.ComputeHash(fullPath)
	if err != nil {
		return false
	}
	if ourHash != resource.Hash {
		slog.DebugContext(ctx, "resource is MISSED. HASH MISMATCH!!",
			"path", fullPath,
			"ourHash", ourHash,
			"resourceHash", resource.Hash)
		return false
	}
	slog.DebugContext(ctx, "resource is cached!!", "path", fullPath)
	return true
}

func (*Client) ComputeHash(fullPath string) (string, error) {
	// Read the file, using CRC32.
	reader, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer reader.Close()
	hash := crc32.NewIEEE()
	if _, copyErr := io.Copy(hash, reader); copyErr != nil {
		return "", fmt.Errorf("failed to copy file: %w", copyErr)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (c *Client) getCatalogURL(ctx context.Context, version string) (string, error) {
	resp, err := c.client.R().SetContext(ctx).Get(fmt.Sprintf(GetAssetsVersionURL, version))
	if err != nil || resp.IsError() {
		return "", fmt.Errorf("failed to send request to get assets version: %w", err)
	}

	var result ServerInfoData
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Get the version check response.
	slog.DebugContext(ctx, "getCatalogURL", "resp", string(resp.Body()))
	versionCheckResp, err := c.versionCheck(ctx, result.ServerInfoDataURL)
	if err != nil {
		return "", fmt.Errorf("failed to get version check response: %w", err)
	}

	// Get the assets URL.
	// Get connection group based on result.DefaultConnectionGroup
	for _, connectionGroup := range versionCheckResp.ConnectionGroups {
		if connectionGroup.Name == result.DefaultConnectionGroup {
			latestVersionIndex := len(connectionGroup.OverrideConnectionGroups) - 1
			return connectionGroup.OverrideConnectionGroups[latestVersionIndex].AddressablesCatalogURLRoot, nil
		}
	}

	return "", fmt.Errorf("failed to get catalog root URL: %w", err)
}

func (c *Client) versionCheck(ctx context.Context, assetsURL string) (*VersionCheckResponse, error) {
	resp, err := c.client.R().SetContext(ctx).Get(assetsURL)
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

	version, err := c.GetLatestVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	catalogURL, err := c.getCatalogURL(ctx, version)
	if err != nil {
		return "", fmt.Errorf("failed to get catalog URL: %w", err)
	}
	c.resourcePath = catalogURL
	return c.resourcePath, nil
}
