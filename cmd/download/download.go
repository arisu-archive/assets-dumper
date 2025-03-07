package download

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/sync/errgroup"

	"github.com/arisu-archive/assets-dumper/cmd"
	"github.com/arisu-archive/assets-dumper/pkg/resources"
)

type command struct {
	cmd  *cobra.Command
	opts options
}

type Command interface {
	Command() *cobra.Command
}

func (c *command) Command() *cobra.Command {
	return c.cmd
}

func NewCommand() Command {
	download := &command{}
	command := &cobra.Command{
		Aliases: []string{"d", "dl"},
		Use:     "download --server [global|japan] --output <path> --filter <glob>",
		Short:   "Download assets from a server",
		Example: "assets-dumper download --server global --output ./output --filter '**/*.png'",
		Args:    cobra.NoArgs,
		RunE:    cmd.RunE("download", download.execute),
	}

	command.Flags().StringVarP(&download.opts.server, "server", "s", "", "server to download from")
	command.Flags().StringVarP(&download.opts.output, "output", "o", "", "path to download to")
	command.Flags().StringVarP(&download.opts.filter, "filter", "f", "**", "glob filter to apply to the download")
	command.Flags().IntVarP(
		&download.opts.maxConcurrency,
		"max-concurrency",
		"c",
		10,
		"maximum number of concurrent downloads",
	)
	if err := command.MarkFlagRequired("server"); err != nil {
		panic(err)
	}
	if err := command.MarkFlagRequired("output"); err != nil {
		panic(err)
	}
	download.cmd = command
	return download
}

func (c *command) execute(cobraCmd *cobra.Command, _ []string) error {
	s := resources.GetServer(c.opts.server)
	if !s.IsValid() {
		return fmt.Errorf("invalid server: %s", c.opts.server)
	}

	client, err := resources.NewClient(s)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	files, err := client.GetResources(cobraCmd.Context(), c.opts.filter)
	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}

	return downloadResources(cobraCmd.Context(), downloadConfig{
		client:         client,
		paths:          files,
		outputDir:      c.opts.output,
		maxConcurrency: c.opts.maxConcurrency,
	})
}

// Create a download configuration struct.
type downloadConfig struct {
	client         resources.Client
	paths          []string
	outputDir      string
	maxConcurrency int
}

func downloadResources(ctx context.Context, config downloadConfig) error {
	p := mpb.New(
		mpb.WithWidth(64),
		mpb.WithRefreshRate(100*time.Millisecond),
	)
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(config.maxConcurrency) // Limit concurrent goroutines
	// Pre-process paths and count files to download
	filesToDownload := []string{}
	for _, resourcePath := range config.paths {
		normalizedPath := strings.ReplaceAll(resourcePath, `\`, "/")
		fullPath := filepath.Join(config.outputDir, normalizedPath)
		if _, err := os.Stat(fullPath); err == nil {
			slog.DebugContext(ctx, "skipping resource", "path", fullPath)
			continue
		}
		filesToDownload = append(filesToDownload, normalizedPath)
	}

	bar := p.AddBar(int64(len(filesToDownload)),
		mpb.PrependDecorators(
			decor.Name("Downloading", decor.WC{C: decor.DindentRight | decor.DextraSpace}),
		),
		mpb.AppendDecorators(
			decor.AverageETA(decor.ET_STYLE_GO, decor.WCSyncWidth),
			decor.Name(" | "),
			decor.CountersNoUnit("%d / %d"),
		),
	)

	for _, normalizedPath := range filesToDownload {
		g.Go(func() error {
			start := time.Now()
			subBar := p.New(1,
				mpb.BarStyle(),
				mpb.BarRemoveOnComplete(),
				mpb.PrependDecorators(decor.Name(normalizedPath)),
			)
			if err := saveResource(ctx, config.client, normalizedPath, config.outputDir); err != nil {
				return err
			}
			subBar.Increment()
			subBar.Wait()
			duration := time.Since(start)
			bar.EwmaIncrInt64(1, duration)
			return nil
		})
	}

	// Wait for all downloads to complete or for an error
	err := g.Wait()
	p.Wait() // Wait for progress bars to finish

	return fmt.Errorf("failed to download resources: %w", err)
}

func saveResource(ctx context.Context, client resources.Client, resourcePath, outputDir string) error {
	normalizedPath := filepath.FromSlash(resourcePath)
	fullPath := filepath.Join(outputDir, normalizedPath)

	data, err := client.GetResource(ctx, resourcePath)
	if err != nil {
		return fmt.Errorf("failed to get resource %s: %w", resourcePath, err)
	}

	err = os.MkdirAll(filepath.Dir(fullPath), 0o755)
	if err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", resourcePath, err)
	}

	err = os.WriteFile(fullPath, data, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write resource %s: %w", resourcePath, err)
	}
	slog.DebugContext(ctx, "wrote resource", "path", fullPath)
	return nil
}
