package extract

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/arisu-archive/assets-dumper/cmd"
	"github.com/arisu-archive/assets-dumper/pkg/extractor"
	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
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
	e := &command{}
	command := &cobra.Command{
		Aliases: []string{"x"},
		Use:     "extract --input <path> --output <path> --server <server>",
		Short:   "Extract the encrypted assets from the input path",
		Example: "assets-dumper x -i ./input -o ./output -s global",
		Args:    cobra.NoArgs,
		RunE:    cmd.RunE("extract", e.execute),
	}

	command.Flags().StringVarP(&e.opts.server, "server", "s", "", "server to extract assets from")
	command.Flags().StringVarP(&e.opts.inputPath, "input", "i", "", "path to the encrypted assets")
	command.Flags().StringVarP(&e.opts.outputPath, "output", "o", "", "path to save the decrypted assets")
	command.Flags().StringVarP(&e.opts.key, "key", "k", "", "sqlcipher database key (required for encrypted Japan databases)")
	if err := command.MarkFlagRequired("input"); err != nil {
		panic(err)
	}
	if err := command.MarkFlagRequired("output"); err != nil {
		panic(err)
	}
	if err := command.MarkFlagRequired("server"); err != nil {
		panic(err)
	}

	e.cmd = command

	return e
}

func (c *command) execute(cobraCmd *cobra.Command, _ []string) error {
	ctx := cobraCmd.Context()
	if err := c.opts.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	s := resourceapi.GetServer(c.opts.server)
	if !s.IsValid() {
		return fmt.Errorf("invalid server: %s", c.opts.server)
	}

	var opts []extractor.Option
	if c.opts.key != "" {
		key, err := base64.StdEncoding.DecodeString(c.opts.key)
		if err != nil {
			return fmt.Errorf("invalid key format: %w", err)
		}
		opts = append(opts, extractor.WithKey(key))
	}

	client := extractor.New(s, opts...)
	if err := client.Extract(ctx, c.opts.inputPath, c.opts.outputPath); err != nil {
		return fmt.Errorf("failed to extract assets: %w", err)
	}

	return nil
}
