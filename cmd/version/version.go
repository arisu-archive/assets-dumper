package version

import (
	"fmt"
	"log/slog"

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

func NewCommand() Command {
	version := &command{}
	command := &cobra.Command{
		Aliases: []string{"v", "ver"},
		Use:     "version --server [global|japan]",
		Short:   "Check the game latest version",
		Example: "assets-dumper version --server [global|japan]",
		Args:    cobra.NoArgs,
		RunE:    cmd.RunE("version", version.execute),
	}
	command.Flags().StringVarP(&version.opts.server, "server", "s", "", "server to check the version from")
	command.Flags().BoolVar(&version.opts.withoutPatch, "without-patch", false, "without patch version")
	version.cmd = command
	return version
}

func (c *command) execute(cobraCmd *cobra.Command, _ []string) error {
	s := resourceapi.GetServer(c.opts.server)
	if !s.IsValid() {
		return fmt.Errorf("invalid server: %s", c.opts.server)
	}

	client, err := resources.NewClient(s)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	version, err := client.GetLatestVersion(cobraCmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}
	slog.Info("latest version", "version", version)
	patchVersion := ""
	if !c.opts.withoutPatch {
		patchVersion, err = client.GetLatestPatchVersion(cobraCmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get patch version: %w", err)
		}
		slog.Info("patch version", "version", patchVersion)
	}

	// Concatenate version and patch version
	fullVersion := fmt.Sprintf("%s-%s", version, patchVersion)
	_, err = cobraCmd.OutOrStdout().Write([]byte(fullVersion + "\n"))
	if err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}
	return nil
}

func (c *command) Command() *cobra.Command {
	return c.cmd
}
