package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:          "godeez",
	Short:        "GoDeez is a tool to download music from Deezer",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// TEMPORARILY DISABLED:
		// Watcher autostart (EnsureAutostart) has been disabled due to
		// concurrency issues with database access (e.g., when using `download`).
		// To re-enable, uncomment the line below.

		/*
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			if err := watcher.EnsureAutostart(homeDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to install autostart for watcher: %v\n", err)
			}
		*/

		return nil
	},
}
