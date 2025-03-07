package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
)

type CommandExecutor func(cmd *cobra.Command, args []string) (err error)

type Command interface {
	Command() *cobra.Command
}

func RunE(verb string, fn func(cmd *cobra.Command, args []string) error) CommandExecutor {
	return func(cmd *cobra.Command, args []string) (err error) {
		start := time.Now()
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%s failed after %s: %v", verb, time.Since(start).Truncate(time.Second), r)
			}
		}()
		if err = fn(cmd, args); err != nil {
			return fmt.Errorf("%s failed after %s: %w", verb, time.Since(start).Truncate(time.Second), err)
		}
		slog.Info(fmt.Sprintf("%s completed after %s", verb, time.Since(start).Truncate(time.Second)))
		return nil
	}
}
