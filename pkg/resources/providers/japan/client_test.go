package japan_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/onsi/gomega/ghttp"

	"github.com/arisu-archive/memorypack-go"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
	"github.com/arisu-archive/assets-dumper/pkg/resources/providers/japan"
)

// Define a custom RoundTripper for url redirects.
type RoundTripFunc func(*http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var _ = Describe("Japan Client", func() {
	var (
		server *ghttp.Server
		client *japan.Client
		ctx    context.Context
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		// Don't set the base URL - instead intercept all requests
		restyClient := resty.New()

		// Use the transport proxy pattern to redirect requests to our test server
		originalTransport := http.DefaultTransport
		restyClient.SetTransport(RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Rewrite URLs to go to our test server
			origURL := req.URL.String()

			// Handle japan version URLs
			if strings.Contains(origURL, "version.txt") {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(server.URL(), "http://")
				req.URL.Path = "/com.YostarJP.BlueArchive/version.txt"
			}

			// Handle version check URL
			if strings.Contains(origURL, "master-version-check") {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(server.URL(), "http://")
				req.URL.Path = "/master-version-check"
			}

			// Handle resource requests
			if strings.Contains(origURL, "images/") || strings.Contains(origURL, "MediaResources/") {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(server.URL(), "http://")
			}

			// Handle assets URL
			if strings.Contains(origURL, "GameMainConfig.json") {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(server.URL(), "http://")
			}

			resp, respErr := originalTransport.RoundTrip(req)
			return resp, respErr
		}))

		client = japan.NewClient(restyClient)
		ctx = context.Background()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("GetVersion", func() {
		It("should retrieve version successfully", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/com.YostarJP.BlueArchive/version.txt"),
					ghttp.RespondWith(http.StatusOK, "1.2.3"),
				),
			)

			version, err := client.GetVersion(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("1.2.3"))
		})

		It("should handle error conditions", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/com.YostarJP.BlueArchive/version.txt"),
					ghttp.RespondWith(http.StatusInternalServerError, "Server Error"),
				),
			)

			_, err := client.GetVersion(ctx)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetResource", func() {
		It("should download resource correctly", func() {
			// Mock version check response
			serverInfoData := japan.ServerInfoData{
				ServerInfoDataURL: server.URL() + "/serverinfo.json",
			}
			serverInfoJSON, _ := json.Marshal(serverInfoData)

			versionCheckResponse := japan.VersionCheckResponse{
				ConnectionGroups: []japan.ConnectionGroup{
					{
						OverrideConnectionGroups: []japan.OverrideConnectionGroup{
							{Name: "Test1"},
							{Name: "Test2", AddressablesCatalogURLRoot: server.URL() + "/resources"},
						},
					},
				},
			}
			versionCheckJSON, _ := json.Marshal(versionCheckResponse)

			// Resource content
			resourceContent := []byte("resource data")

			// Set up server responses
			server.AppendHandlers(
				// Version check
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/com.YostarJP.BlueArchive/version.txt"),
					ghttp.RespondWith(http.StatusOK, "1.2.3"),
				),
				// Assets URL
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/com.YostarJP.BlueArchive/decompiled/1.2.3/GameMainConfig.json"),
					ghttp.RespondWith(http.StatusOK, string(serverInfoJSON)),
				),
				// Version check API call
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/serverinfo.json"),
					ghttp.RespondWith(http.StatusOK, string(versionCheckJSON)),
				),
				// Resource download
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/resources/images/icon1.png"),
					ghttp.RespondWith(http.StatusOK, resourceContent),
				),
			)

			reader, size, err := client.DownloadResource(ctx, "images/icon1.png")
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(int64(len(resourceContent))))
			data, err := io.ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(resourceContent))
		})
	})

	Describe("GetResources", func() {
		It("should filter resources correctly", func() {
			// Mock version check response
			serverInfoData := japan.ServerInfoData{
				ServerInfoDataURL: server.URL() + "/serverinfo.json",
			}
			serverInfoJSON, _ := json.Marshal(serverInfoData)

			versionCheckResponse := japan.VersionCheckResponse{
				ConnectionGroups: []japan.ConnectionGroup{
					{
						OverrideConnectionGroups: []japan.OverrideConnectionGroup{
							{Name: "Test1"},
							{Name: "Test2", AddressablesCatalogURLRoot: server.URL() + "/resources"},
						},
					},
				},
			}
			versionCheckJSON, _ := json.Marshal(versionCheckResponse)

			// Create mock catalog data
			tableCatalog := japan.TableCatalog{
				TableBundles: map[string]japan.TableBundle{
					"table1": {Path: "table1.bytes"},
					"table2": {Path: "table2.bytes"},
				},
			}
			tableCatalogBytes, _ := memorypack.Serialize(&tableCatalog)

			mediaCatalog := japan.MediaCatalog{
				MediaBundles: map[string]japan.MediaBundle{
					"media1": {Path: "media/image1.png"},
					"media2": {Path: "media/audio/sound1.mp3"},
				},
			}
			mediaCatalogBytes, _ := memorypack.Serialize(&mediaCatalog)

			bundleDownloadInfo := japan.BundleDownloadInfo{
				Files: []japan.BundleFile{
					{Name: "bundle1.bundle"},
					{Name: "bundle2.bundle"},
				},
			}
			bundleDownloadInfoJSON, _ := json.Marshal(bundleDownloadInfo)

			// Set up server responses
			server.AppendHandlers(
				// Version check
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/com.YostarJP.BlueArchive/version.txt"),
					ghttp.RespondWith(http.StatusOK, "1.2.3"),
				),
				// Assets URL
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/com.YostarJP.BlueArchive/decompiled/1.2.3/GameMainConfig.json"),
					ghttp.RespondWith(http.StatusOK, string(serverInfoJSON)),
				),
				// Version check API call
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/serverinfo.json"),
					ghttp.RespondWith(http.StatusOK, string(versionCheckJSON)),
				),
				// TableCatalog request
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/resources/TableBundles/TableCatalog.bytes"),
					ghttp.RespondWith(http.StatusOK, tableCatalogBytes),
				),
				// MediaCatalog request
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/resources/MediaResources/Catalog/MediaCatalog.bytes"),
					ghttp.RespondWith(http.StatusOK, mediaCatalogBytes),
				),
				// BundleDownloadInfo request
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/resources/Android/bundleDownloadInfo.json"),
					ghttp.RespondWith(http.StatusOK, bundleDownloadInfoJSON),
				),
			)

			// Filter for only media files
			resources, err := client.ListResources(ctx, "MediaResources/**")
			Expect(err).NotTo(HaveOccurred())
			Expect(resources).To(HaveLen(2))
			Expect(resources).To(ContainElements(
				resourceapi.Resource{Path: "MediaResources/media/image1.png", Size: 0, Hash: "0"},
				resourceapi.Resource{Path: "MediaResources/media/audio/sound1.mp3", Size: 0, Hash: "0"},
			))
		})
	})
})
