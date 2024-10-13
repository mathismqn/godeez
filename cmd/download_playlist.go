package cmd

import "github.com/spf13/cobra"

var playlistCmd = &cobra.Command{
	Use:   "playlist [playlist_id...]",
	Short: "Download songs from one or more playlists",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateInput()
		downloadContent("playlist", args)
	},
}

func init() {
	downloadCmd.AddCommand(playlistCmd)
}
