package global_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
	"github.com/arisu-archive/assets-dumper/pkg/resources/providers/global"
)

// Define a custom RoundTripper for url redirects.
type RoundTripFunc func(*http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var _ = Describe("Global Client", func() {
	var (
		server *ghttp.Server
		client *global.Client
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

			// Handle global.GetVersionURL
			if strings.Contains(origURL, "version.txt") {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(server.URL(), "http://")
				req.URL.Path = "/com.nexon.bluearchive/version.txt"
			}

			// Handle global.VersionCheckURL
			if strings.Contains(origURL, "version-check") {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(server.URL(), "http://")
				req.URL.Path = "/patch/v1.1/version-check"
			}

			// Handle resource requests
			if strings.Contains(origURL, "images/") {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(server.URL(), "http://")
				// Keep the original path
			}

			return originalTransport.RoundTrip(req)
		}),
		)

		client = global.NewClient(restyClient)
		ctx = context.Background()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("GetVersion", func() {
		It("should retrieve version successfully", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/com.nexon.bluearchive/version.txt"),
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
					ghttp.VerifyRequest("GET", "/com.nexon.bluearchive/version.txt"),
					ghttp.RespondWith(http.StatusInternalServerError, "Server Error"),
				),
			)

			_, err := client.GetVersion(ctx)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetResources", func() {
		It("should filter resources correctly", func() {
			// Mock version check response
			versionCheckResponse := global.VersionCheckResponse{
				APIVersion:         "1.1",
				LatestBuildVersion: "1.2.3",
				Patch: global.Patch{
					PatchVersion: 123,
					ResourcePath: server.URL() + "/resources.json",
				},
			}
			versionCheckJSON, _ := json.Marshal(versionCheckResponse)

			// Mock resource data
			resourceData := global.ResourceData{
				Resources: []global.Resource{
					{ResourcePath: "images/icon1.png", ResourceSize: 100},
					{ResourcePath: "images/icon2.png", ResourceSize: 200},
					{ResourcePath: "audio/sound1.mp3", ResourceSize: 300},
				},
			}
			resourceJSON, _ := json.Marshal(resourceData)

			// Set up server responses
			server.AppendHandlers(
				// Version check
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/com.nexon.bluearchive/version.txt"),
					ghttp.RespondWith(http.StatusOK, "1.2.3"),
				),
				// Version check API call
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/patch/v1.1/version-check"),
					ghttp.RespondWith(http.StatusOK, string(versionCheckJSON)),
				),
				// Resources list
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/resources.json"),
					ghttp.RespondWith(http.StatusOK, string(resourceJSON)),
				),
			)

			// Get resources with a filter that matches only images
			resources, err := client.ListResources(ctx, "images/**")
			Expect(err).NotTo(HaveOccurred())
			Expect(resources).To(HaveLen(2))
			Expect(resources).To(ContainElements(
				resourceapi.Resource{Path: "images/icon1.png", Size: 100},
				resourceapi.Resource{Path: "images/icon2.png", Size: 200},
			))
		})
	})

	Describe("GetResource", func() {
		It("should download resource correctly", func() {
			// Mock version check response
			versionCheckResponse := global.VersionCheckResponse{
				APIVersion:         "1.1",
				LatestBuildVersion: "1.2.3",
				Patch: global.Patch{
					PatchVersion: 123,
					ResourcePath: server.URL() + "/resources.json",
				},
			}
			versionCheckJSON, _ := json.Marshal(versionCheckResponse)

			// Resource content
			resourceContent := []byte("resource data")

			// Set up server responses
			server.AppendHandlers(
				// Version check
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/com.nexon.bluearchive/version.txt"),
					ghttp.RespondWith(http.StatusOK, "1.2.3"),
				),
				// Version check API call
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/patch/v1.1/version-check"),
					ghttp.RespondWith(http.StatusOK, string(versionCheckJSON)),
				),
				// Resource download
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/images/icon1.png"),
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
})
