package global

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"path"
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
	resourcePath string
	resourceURI  string
}

func NewClient(client *resty.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) ListResources(ctx context.Context, filter string) ([]resourceapi.Resource, error) {
	if _, err := c.getResourceURI(ctx); err != nil {
		return nil, fmt.Errorf("failed to get resource path: %w", err)
	}

	// Fetch the resources from the resource path
	resp, err := c.client.R().SetContext(ctx).Get(c.resourcePath)
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
	resp, err := c.client.R().SetDoNotParseResponse(true).SetContext(ctx).Get(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download resource: %w", err)
	}
	return resp.RawBody(), resp.RawResponse.ContentLength, nil
}

func (c *Client) GetVersion(ctx context.Context) (string, error) {
	resp, err := c.client.R().SetContext(ctx).Get(GetVersionURL)
	if err != nil || resp.IsError() {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	slog.DebugContext(ctx, "getVersion", "resp", resp.Body())
	return strings.TrimSpace(string(resp.Body())), nil
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
	slog.DebugContext(ctx, "versionCheck", "resp", resp.Body())

	if resp.IsError() {
		return nil, fmt.Errorf("failed to perform version check: %v", resp.Body())
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
	// TODO: Support incremental update.
	parsedURL, err := url.Parse(versionCheckResp.Patch.ResourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}
	slog.DebugContext(ctx, "getResourceURI", "parsedURL", parsedURL)
	// Remove the filename from the path
	dir, _ := path.Split(parsedURL.Path)
	parsedURL.Path = dir
	c.resourcePath = versionCheckResp.Patch.ResourcePath
	c.resourceURI = strings.TrimSuffix(parsedURL.String(), "/")
	return c.resourceURI, nil
}
