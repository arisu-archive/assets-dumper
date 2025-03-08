package root

import (
	"io"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/arisu-archive/assets-dumper/cmd/catalog"
	"github.com/arisu-archive/assets-dumper/cmd/download"
	"github.com/arisu-archive/assets-dumper/cmd/version"
)

type command struct {
	cmd     *cobra.Command
	exit    func(int)
	verbose bool
}

// ExecuteConfig holds the configuration for the root command.
type ExecuteConfig struct {
	Version string
	Exit    func(int)
	In      io.Reader
	Out     io.Writer
	Err     io.Writer
	Args    []string
}

func Execute(config ExecuteConfig) {
	newRootCmd(config).Execute(config.Args)
}

func (r *command) Execute(args []string) {
	r.cmd.SetArgs(args)
	if err := r.cmd.Execute(); err != nil {
		slog.Error("assets-dumper failed.", slog.Any("error", err))
		r.exit(1)
	}
}

func newRootCmd(config ExecuteConfig) *command {
	root := &command{
		exit: config.Exit,
	}

	root.cmd = &cobra.Command{
		Use:   "assets-dumper <command> [flags]",
		Short: "Download assets from the game.",
		Long: `assets-dumper is a tool to download assets from the game.
You can download assets from the global or japan server.`,
		Example:           `assets-dumper download --server global --output ./assets`,
		Version:           config.Version,
		SilenceErrors:     false,
		SilenceUsage:      false,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if root.verbose {
				slog.SetDefault(slog.New(slog.NewTextHandler(config.Out, &slog.HandlerOptions{Level: slog.LevelDebug})))
				slog.Debug("verbose mode enabled")
			}
			return nil
		},
		PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
			slog.Info("assets-dumper finished")
			return nil
		},
	}

	root.cmd.PersistentFlags().BoolVarP(&root.verbose, "verbose", "v", false, "verbose mode")
	root.cmd.AddCommand(download.NewCommand().Command())
	root.cmd.AddCommand(version.NewCommand().Command())
	root.cmd.AddCommand(catalog.NewCommand().Command())
	root.cmd.SetIn(config.In)
	root.cmd.SetOut(config.Out)
	root.cmd.SetErr(config.Err)

	return root
}
