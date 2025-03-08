package download

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/arisu-archive/assets-dumper/cmd"
	"github.com/arisu-archive/assets-dumper/pkg/downloader"
	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
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
	ctx := cobraCmd.Context()
	s := resourceapi.GetServer(c.opts.server)
	if !s.IsValid() {
		return fmt.Errorf("invalid server: %s", c.opts.server)
	}

	client, err := resources.NewClient(s)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	files, err := client.ListResources(ctx, c.opts.filter)
	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
	}
	filesToDownload := c.getFilteredFiles(ctx, client, files)
	if len(filesToDownload) == 0 {
		slog.DebugContext(ctx, "no resources to download")
		return nil
	}

	d := downloader.New(client, c.opts.maxConcurrency)
	if downloadErr := d.DownloadAll(ctx, filesToDownload, c.opts.output); downloadErr != nil {
		return fmt.Errorf("failed to download resources: %w", downloadErr)
	}
	return nil
}

func (c *command) getFilteredFiles(
	ctx context.Context,
	client resourceapi.Client,
	files []resourceapi.Resource,
) []string {
	filesToDownload := []string{}
	for _, resource := range files {
		normalizedPath := strings.ReplaceAll(resource.Path, `\`, "/")
		fullPath := filepath.Join(c.opts.output, normalizedPath)
		if !client.IsResourceCached(ctx, resource, fullPath) {
			filesToDownload = append(filesToDownload, normalizedPath)
		}
	}
	return filesToDownload
}
