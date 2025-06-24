package cmd

import (
	"fmt"
	"os"

	"github.com/mathismqn/godeez/internal/watcher"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:          "godeez",
	Short:        "GoDeez is a tool to download music from Deezer",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		if err := watcher.EnsureAutostart(homeDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to install autostart for watcher: %v\n", err)
		}

		return nil
	},
}
