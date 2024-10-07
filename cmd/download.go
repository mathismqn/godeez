package cmd

import (
	"fmt"
	"os"

	"github.com/mathismqn/godeez/internal/album"
	"github.com/mathismqn/godeez/internal/song"
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
			albumData, err := album.GetDataByID(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not fetch album data: %v\n", err)
				continue
			}

			for _, s := range albumData.Songs.Data {
				media, err := song.GetMedia(s)
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
