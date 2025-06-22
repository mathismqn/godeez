package cmd

import (
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch playlists and auto-download new tracks",
}

func init() {
	RootCmd.AddCommand(watchCmd)
}
