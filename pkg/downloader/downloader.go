package downloader

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/sync/errgroup"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

// Downloader is a struct that contains a client and a max concurrency.
type Downloader struct {
	client         resourceapi.Client
	maxConcurrency int
}

// New creates a new Downloader.
func New(client resourceapi.Client, maxConcurrency int) *Downloader {
	return &Downloader{
		client:         client,
		maxConcurrency: maxConcurrency,
	}
}

// DownloadAll downloads all resources in parallel.
func (d *Downloader) DownloadAll(ctx context.Context, files []string, outputDir string) error {
	p := mpb.New(
		mpb.WithWidth(64),
		mpb.WithRefreshRate(100*time.Millisecond),
	)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(d.maxConcurrency)
	for _, normalizedPath := range files {
		path := normalizedPath // Create a copy for the closure
		g.Go(func() error {
			return d.download(ctx, path, outputDir, p)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to download resources: %w", err)
	}
	return nil
}

// download downloads a resource and writes it to the output directory.
func (d *Downloader) download(ctx context.Context, resourcePath, outputDir string, p *mpb.Progress) error {
	fullPath := filepath.Join(outputDir, resourcePath)

	reader, size, err := d.client.DownloadResource(ctx, resourcePath)
	if err != nil {
		return fmt.Errorf("failed to get resource %s: %w", resourcePath, err)
	}
	defer reader.Close()

	if makeErr := os.MkdirAll(filepath.Dir(fullPath), 0o755); makeErr != nil {
		return fmt.Errorf("failed to create directory for %s: %w", resourcePath, makeErr)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fullPath, err)
	}
	defer file.Close()

	bar := createProgressBar(p, resourcePath, size)
	if _, writeErr := io.Copy(bar.ProxyWriter(file), reader); writeErr != nil {
		bar.Abort(true)
		return fmt.Errorf("failed to write resource %s: %w", resourcePath, writeErr)
	}

	if size == 0 {
		bar.SetTotal(0, true)
	}
	bar.Wait()
	slog.DebugContext(ctx, "wrote resource", "path", fullPath)
	return nil
}

// createProgressBar creates a progress bar for a resource download.
func createProgressBar(p *mpb.Progress, resourcePath string, size int64) *mpb.Bar {
	return p.AddBar(size,
		mpb.BarFillerClearOnComplete(),
		mpb.PrependDecorators(
			decor.OnComplete(
				decor.Name(fmt.Sprintf("Downloading: %s", resourcePath), decor.WC{C: decor.DindentRight}),
				fmt.Sprintf("Completed: %s", resourcePath),
			),
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO),
				"",
			),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
		),
	)
}
