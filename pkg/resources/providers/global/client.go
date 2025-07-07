package global

import (
	"context"
	"crypto/md5" //nolint:gosec //False positive: Using MD5 is not our choice.
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

var _ resourceapi.Client = (*Client)(nil)

type Client struct {
	client       *resty.Client
	retriever    *CatalogRetriever
	resourcePath string
	resourceURI  string
	version      string
	patchVersion int64
	patchPath    string
	patchURI     string
}

func NewClient(client *resty.Client) *Client {
	c := &Client{
		client: client,
	}
	c.retriever = NewCatalogRetriever(c)
	return c
}

func NewClientWithRetriever(client *resty.Client, retriever *CatalogRetriever) *Client {
	c := &Client{
		client:    client,
		retriever: retriever,
	}
	return c
}

func (c *Client) WithVersion(version string) resourceapi.Client {
	c.version = version
	c.resourceURI = ""
	c.resourcePath = ""
	c.patchVersion = 0
	c.patchPath = ""
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

func (c *Client) ListResources(ctx context.Context, filter string) ([]resourceapi.Resource, error) {
	if _, err := c.getResourceURI(ctx); err != nil {
		return nil, fmt.Errorf("failed to get resource path: %w", err)
	}

	// Fetch the resources from the resource path
	resp, err := c.client.R().SetContext(ctx).Get(c.resourcePath)
	slog.DebugContext(ctx, "listResources", "resp", resp.Body())
	if err != nil {
		return nil, fmt.Errorf("failed to download resource: %w", err)
	}
	var result ResourceData
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	resources := []resourceapi.Resource{}
	for _, resource := range result.Resources {
		slog.DebugContext(ctx, "resource", "resource", resource.ResourcePath)
		// Given a list of folder paths, we need to filter out the ones that don't match the filter using glob
		matches, matchErr := doublestar.Match(filter, resource.ResourcePath)
		if matchErr != nil {
			slog.ErrorContext(ctx, "failed to match resource", "error", matchErr)
			continue
		}
		if matches {
			resources = append(resources, resourceapi.Resource{
				Path: resource.ResourcePath,
				Size: resource.ResourceSize,
				Hash: resource.ResourceHash,
			})
		}
	}

	return resources, nil
}

func (c *Client) DownloadResource(ctx context.Context, resourcePath string) (io.ReadCloser, int64, error) {
	resourceURI, err := c.getResourceURI(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get resource path: %w", err)
	}
	fullPath := fmt.Sprintf("%s/%s", resourceURI, strings.TrimPrefix(resourcePath, "/"))
	slog.DebugContext(ctx, "DownloadResource", "fullPath", fullPath)
	resp, err := c.client.R().SetDoNotParseResponse(true).SetContext(ctx).Get(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download resource: %w", err)
	}
	return resp.RawBody(), resp.RawResponse.ContentLength, nil
}

func (c *Client) DownloadPatch(ctx context.Context, patchPath string) (io.ReadCloser, int64, error) {
	_, err := c.getResourceURI(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get resource path: %w", err)
	}
	fullPath := fmt.Sprintf("%s/%s", c.patchURI, strings.TrimPrefix(patchPath, "/"))
	slog.DebugContext(ctx, "DownloadPatch", "fullPath", fullPath)
	resp, err := c.client.R().SetDoNotParseResponse(true).SetContext(ctx).Get(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download patch: %w", err)
	}
	return resp.RawBody(), resp.RawResponse.ContentLength, nil
}

func (c *Client) DownloadResourceToFile(ctx context.Context, resourcePath string) ([]byte, error) {
	resourceURI, err := c.getResourceURI(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource path: %w", err)
	}
	fullPath := fmt.Sprintf("%s/%s", resourceURI, strings.TrimPrefix(resourcePath, "/"))
	slog.DebugContext(ctx, "DownloadResourceToFile", "fullPath", fullPath)
	resp, err := c.client.R().SetContext(ctx).Get(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download resource: %w", err)
	}
	slog.DebugContext(ctx, "DownloadResourceToFile", "resp", resp.Body()[:100])
	return resp.Body(), nil
}

func (c *Client) GetVersion(ctx context.Context) (string, error) {
	// Return the pass-in version if it's not empty.
	if c.version != "" {
		return c.version, nil
	}

	resp, err := c.client.R().SetContext(ctx).Get(GetVersionURL)
	if err != nil || resp.IsError() {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	slog.DebugContext(ctx, "getVersion", "resp", resp.Body())
	return strings.TrimSpace(string(resp.Body())), nil
}

func (c *Client) GetPatchVersion(ctx context.Context) (string, error) {
	if _, err := c.getResourceURI(ctx); err != nil {
		return "", fmt.Errorf("failed to get resource path: %w", err)
	}

	return strconv.FormatInt(c.patchVersion, 10), nil
}

func (c *Client) ListPatches(ctx context.Context, filter string) ([]resourceapi.Resource, error) {
	if _, err := c.getResourceURI(ctx); err != nil {
		return nil, fmt.Errorf("failed to get resource path: %w", err)
	}

	resp, err := c.client.R().SetContext(ctx).Get(c.patchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download patch: %w", err)
	}
	slog.DebugContext(ctx, "listPatches", "resp", resp.Body())
	var result []string
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	patches := []resourceapi.Resource{}
	for _, patch := range result {
		patches = append(patches, resourceapi.Resource{
			Path: patch,
		})
	}
	return patches, nil
}

func (c *Client) IsResourceCached(_ context.Context, resource resourceapi.Resource, fullPath string) bool {
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
		return false
	}
	return true
}

func (*Client) ComputeHash(fullPath string) (string, error) {
	// Read the file, using MD5.
	reader, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer reader.Close()
	hash := md5.New() //nolint:gosec //False positive: Using MD5 is not our choice.
	if _, copyErr := io.Copy(hash, reader); copyErr != nil {
		return "", fmt.Errorf("failed to copy file: %w", copyErr)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (c *Client) versionCheck(ctx context.Context) (*VersionCheckResponse, error) {
	buildVersion, err := c.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	buildNumber, err := semver.NewVersion(buildVersion)
	if err != nil {
		return nil, ErrInvalidBuildVersion
	}

	// country := Locale.getDefault().getCountry()
	// language := Locale.getDefault().getLanguage()
	// sdk_version is hardcoded in app source code.
	resp, err := c.client.R().
		SetBody(map[string]any{
			"advertising_id":     uuid.New().String(),
			"country":            DefaultCountry,
			"curr_build_number":  buildNumber.Patch(),
			"curr_build_version": buildVersion,
			"sdk_version":        SdkVersion,
			"curr_patch_version": DefaultPatchVersion,
			"language":           DefaultLanguage,
			"market_code":        MarketCodePlayStore,
			"market_game_id":     AppBundleID,
		}).
		SetHeaders(map[string]string{
			"Content-Type": "application/json; charset=utf-8",
			"User-Agent":   "Dalvik/2.1.0 (Linux; U; Android 14; Pixel 7 Build/TP1A.221005.003)",
		}).
		Post(VersionCheckURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	slog.DebugContext(ctx, "versionCheck", "resp", string(resp.Body()))

	if resp.IsError() {
		return nil, fmt.Errorf("failed to perform version check: %v", string(resp.Body()))
	}

	var result VersionCheckResponse
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &result, nil
}

func (c *Client) getResourceURI(ctx context.Context) (string, error) {
	if c.resourceURI != "" {
		return c.resourceURI, nil
	}

	versionCheckResp, err := c.versionCheck(ctx)
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(versionCheckResp.Patch.ResourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}
	slog.DebugContext(ctx, "getResourceURI", "parsedURL", parsedURL)
	// Remove the filename from the path
	dir, _ := path.Split(parsedURL.Path)
	parsedURL.Path = dir
	c.resourcePath = versionCheckResp.Patch.ResourcePath
	c.patchVersion = versionCheckResp.Patch.PatchVersion
	// Get the value of the map[string]string (bdiffPath is the map and we dont know the key)
	for _, v := range versionCheckResp.Patch.BdiffPath[0] {
		parsedPatchURL, err := url.Parse(v)
		if err != nil {
			return "", fmt.Errorf("failed to parse URL: %w", err)
		}
		c.patchPath = v
		dir, _ := path.Split(parsedPatchURL.Path)
		parsedPatchURL.Path = dir
		c.patchURI = strings.TrimSuffix(parsedPatchURL.String(), "/")
		break
	}

	c.resourceURI = strings.TrimSuffix(parsedURL.String(), "/")
	return c.resourceURI, nil
}
