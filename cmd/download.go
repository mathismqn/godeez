package cmd

import "github.com/spf13/cobra"

var downloadCmd = &cobra.Command{
	Use:   "download [ID]",
	Short: "Download a track from Deezer",
	Run: func(cmd *cobra.Command, args []string) {
		// Your code here
	},
}

func init() {
	RootCmd.AddCommand(downloadCmd)
}
