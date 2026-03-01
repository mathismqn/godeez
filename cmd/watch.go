package cmd

import (
	"github.com/spf13/cobra"
)

// NOTE: The watch command is disabled due to database concurrency issues.
// To re-enable, uncomment RootCmd.AddCommand(watchCmd) in init().
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch playlists and auto-download new tracks",
}

func init() {
	// RootCmd.AddCommand(watchCmd)
}
