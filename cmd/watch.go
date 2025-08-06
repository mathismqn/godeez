package cmd

import (
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch playlists and auto-download new tracks",
}

// TEMPORARILY DISABLED:
// The `watch` command and all its subcommands are currently disabled
// due to known issues (e.g., database access conflicts with `download`).
// To re-enable, uncomment the line below.
func init() {
	// RootCmd.AddCommand(watchCmd)
}
