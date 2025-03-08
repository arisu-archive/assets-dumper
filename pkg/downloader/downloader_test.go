package downloader_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mocks "github.com/arisu-archive/assets-dumper/mocks/github.com/arisu-archive/assets-dumper/pkg/resourceapi"
	"github.com/arisu-archive/assets-dumper/pkg/downloader"
)

var _ = Describe("Downloader", func() {
	var (
		mockClient *mocks.MockClient
		dl         *downloader.Downloader
		tempDir    string
		ctx        context.Context
		cancelFunc context.CancelFunc
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "downloader-test-*")
		Expect(err).NotTo(HaveOccurred())

		mockClient = new(mocks.MockClient)
		ctx, cancelFunc = context.WithCancel(context.Background()) //nolint:fatcontext //False positive
	})

	AfterEach(func() {
		cancelFunc()
		os.RemoveAll(tempDir)
	})

	Describe("New", func() {
		It("creates a new Downloader with the given client and max concurrency", func() {
			dl = downloader.New(mockClient, 5)
			Expect(dl).NotTo(BeNil())
		})
	})

	Describe("DownloadAll", func() {
		BeforeEach(func() {
			dl = downloader.New(mockClient, 3)
		})

		Context("with a successful download", func() {
			BeforeEach(func() {
				mockClient.On("DownloadResource", mock.Anything, mock.Anything).
					Return(func(context.Context, string) io.ReadCloser {
						return io.NopCloser(strings.NewReader("test data"))
					}, int64(9), nil)
			})

			It("downloads all files", func() {
				files := []string{"file1.txt", "dir/file2.txt"}
				err := dl.DownloadAll(ctx, files, tempDir)
				Expect(err).NotTo(HaveOccurred())

				// Verify files were created
				for _, file := range files {
					fullPath := filepath.Join(tempDir, file)
					Expect(fullPath).To(BeARegularFile())

					content, readErr := os.ReadFile(fullPath)
					Expect(readErr).NotTo(HaveOccurred())
					Expect(string(content)).To(Equal("test data"))
				}

				mockClient.AssertNumberOfCalls(GinkgoT(), "DownloadResource", len(files))
			})

			It("handles empty file list", func() {
				err := dl.DownloadAll(ctx, []string{}, tempDir)
				Expect(err).NotTo(HaveOccurred())
				mockClient.AssertNotCalled(GinkgoT(), "DownloadResource")
			})
		})

		Context("with a download error", func() {
			BeforeEach(func() {
				mockClient.On("DownloadResource", mock.Anything, mock.Anything).
					Return(nil, int64(0), errors.New("download error"))
			})

			It("returns an error", func() {
				err := dl.DownloadAll(ctx, []string{"file1.txt"}, tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to download resources"))
			})
		})

		Context("with a context cancellation", func() {
			It("stops downloading when context is canceled", func() {
				mockClient.On("DownloadResource", mock.Anything, mock.Anything).
					Run(func(mock.Arguments) {
						// Cancel the context when this is called
						cancelFunc()
					}).
					Return(nil, int64(0), context.Canceled)

				err := dl.DownloadAll(ctx, []string{"file1.txt"}, tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("context canceled"))
			})
		})

		Context("with directory creation error", func() {
			BeforeEach(func() {
				// Create a file with the same name as a directory we'll try to create
				fileAsDir := filepath.Join(tempDir, "blocked-dir")
				err := os.WriteFile(fileAsDir, []byte("test"), 0o644)
				Expect(err).NotTo(HaveOccurred())

				mockClient.On("DownloadResource", mock.Anything, mock.Anything).
					Return(func(context.Context, string) io.ReadCloser {
						return io.NopCloser(strings.NewReader("test data"))
					}, int64(9), nil)
			})

			It("returns an error when directory creation fails", func() {
				err := dl.DownloadAll(ctx, []string{"blocked-dir/file.txt"}, tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create directory"))
			})
		})

		Context("with file creation error", func() {
			BeforeEach(func() {
				// Create a directory with the same name as a file we'll try to create
				makeErr := os.MkdirAll(filepath.Join(tempDir, "file1.txt"), 0o755)
				Expect(makeErr).NotTo(HaveOccurred())

				mockClient.On("DownloadResource", mock.Anything, mock.Anything).
					Return(func(context.Context, string) io.ReadCloser {
						return io.NopCloser(strings.NewReader("test data"))
					}, int64(9), nil)
			})

			It("returns an error when file creation fails", func() {
				err := dl.DownloadAll(ctx, []string{"file1.txt"}, tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create file"))
			})
		})

		Context("with file write error", func() {
			BeforeEach(func() {
				mockClient.On("DownloadResource", mock.Anything, mock.Anything).
					Return(io.NopCloser(&errorReader{err: errors.New("write error")}), int64(9), nil)
			})

			It("returns an error when writing to file fails", func() {
				err := dl.DownloadAll(ctx, []string{"file1.txt"}, tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to write resource"))
			})
		})

		Context("with zero-sized files", func() {
			BeforeEach(func() {
				mockClient.On("DownloadResource", mock.Anything, mock.Anything).
					Return(io.NopCloser(strings.NewReader("")), int64(0), nil)
			})

			It("handles zero-sized files correctly", func() {
				err := dl.DownloadAll(ctx, []string{"empty.txt"}, tempDir)
				Expect(err).NotTo(HaveOccurred())

				fullPath := filepath.Join(tempDir, "empty.txt")
				info, err := os.Stat(fullPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Size()).To(Equal(int64(0)))
			})
		})

		Context("with large files", func() {
			BeforeEach(func() {
				largeData := strings.Repeat("A", 100000)
				mockClient.On("DownloadResource", mock.Anything, mock.Anything).
					Return(func(context.Context, string) io.ReadCloser {
						return io.NopCloser(strings.NewReader(largeData))
					}, int64(len(largeData)), nil)
			})

			It("handles large files correctly", func() {
				err := dl.DownloadAll(ctx, []string{"large.txt"}, tempDir)
				Expect(err).NotTo(HaveOccurred())

				fullPath := filepath.Join(tempDir, "large.txt")
				Expect(fullPath).To(BeARegularFile())

				info, err := os.Stat(fullPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Size()).To(Equal(int64(100000)))
			})
		})

		Context("with many concurrent downloads", func() {
			It("respects max concurrency", func() {
				// Create a lot of files to download
				files := make([]string, 20)
				for i := range 20 {
					files[i] = fmt.Sprintf("file%d.txt", i)
				}

				var downloadCount int32 = 0
				var maxConcurrent int32 = 0
				var mu sync.Mutex

				mockClient.On("DownloadResource", mock.Anything, mock.Anything).
					Run(func(mock.Arguments) {
						atomic.AddInt32(&downloadCount, 1)

						mu.Lock()
						current := atomic.LoadInt32(&downloadCount)
						if current > maxConcurrent {
							maxConcurrent = current
						}
						mu.Unlock()

						// Simulate some work
						time.Sleep(50 * time.Millisecond)
					}).
					Return(func(context.Context, string) io.ReadCloser {
						defer atomic.AddInt32(&downloadCount, -1)
						return io.NopCloser(strings.NewReader("test data"))
					}, int64(9), nil)

				// Set max concurrency to 5
				dl = downloader.New(mockClient, 5)
				err := dl.DownloadAll(ctx, files, tempDir)
				Expect(err).NotTo(HaveOccurred())

				// Max concurrent downloads should respect the limit
				Expect(maxConcurrent).To(BeNumerically("<=", 5))
			})
		})
	})
})
