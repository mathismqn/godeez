package cmd

import (
	"fmt"
	"os"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [album_id...]",
	Short: "Download songs from one or more albums",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		for _, id := range args {
			album, err := deezer.GetAlbumData(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not get album data: %v\n", err)
				continue
			}

			for _, song := range album.Songs.Data {
				media, err := song.GetMediaData()
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not fetch media data: %v\n", err)
					continue
				}

				fmt.Println(media)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(downloadCmd)
}
