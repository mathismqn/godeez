package cmd

import "github.com/spf13/cobra"

var albumCmd = &cobra.Command{
	Use:   "album [album_id...]",
	Short: "Download songs from one or more albums",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateInput()
		downloadContent("album", args)
	},
}

func init() {
	downloadCmd.AddCommand(albumCmd)
}
