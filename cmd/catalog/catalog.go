package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/arisu-archive/assets-dumper/cmd"
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
	c := &command{}
	command := &cobra.Command{
		Aliases: []string{"c"},
		Use:     "catalog --server <jp|gl> --output <path>",
		Short:   "Retrieve the catalog of the assets",
		Example: "assets-dumper c -s jp -o ./output",
		Args:    cobra.NoArgs,
		RunE:    cmd.RunE("catalog", c.execute),
	}

	command.Flags().StringVarP(&c.opts.server, "server", "s", "", "server to retrieve the catalog from")
	command.Flags().StringVarP(&c.opts.output, "output", "o", "", "path to save the catalog")
	command.Flags().StringVarP(&c.opts.version, "version", "", "", "version to retrieve the catalog from")
	command.Flags().StringVarP(&c.opts.version, "ver", "", "", "version to retrieve the catalog from (alias)")
	if err := command.MarkFlagRequired("server"); err != nil {
		panic(err)
	}
	if err := command.MarkFlagRequired("output"); err != nil {
		panic(err)
	}
	c.cmd = command
	return c
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

	bundles := []resourceapi.CatalogType{
		resourceapi.CatalogTypeTableBundle,
		resourceapi.CatalogTypeMediaResources,
		resourceapi.CatalogTypeBundleDownloadInfo,
	}

	// Loop through all the bundles types.
	for _, bundle := range bundles {
		catalog, getErr := client.WithVersion(c.opts.version).GetCatalog(ctx, bundle)
		if getErr != nil {
			return fmt.Errorf("failed to get resource: %w", getErr)
		}

		// Dump the catalog to the output path
		if makeErr := os.MkdirAll(c.opts.output, 0o755); makeErr != nil {
			return fmt.Errorf("failed to create output directory: %w", makeErr)
		}

		// Write the catalog to the output path
		payload, encodeErr := json.Marshal(catalog)
		if encodeErr != nil {
			return fmt.Errorf("failed to marshal catalog: %w", encodeErr)
		}
		fullPath := filepath.Join(c.opts.output, fmt.Sprintf("%s.json", bundle))
		if writeErr := os.WriteFile(fullPath, payload, 0o600); writeErr != nil {
			return fmt.Errorf("failed to write catalog: %w", writeErr)
		}
	}

	return nil
}
