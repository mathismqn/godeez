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
					fmt.Fprintf(os.Stderr, "could not get media data: %v\n", err)
					if err.Error() == "invalid license token" {
						os.Exit(1)
					}
					continue
				}

				songTitle := song.Title
				if song.Version != "" {
					songTitle = fmt.Sprintf("%s %s", song.Title, song.Version)
				}

				filename := fmt.Sprintf("%s - %s.flac", song.ArtistName, songTitle)
				err = media.Download(filename, song.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not download song: %v\n", err)
					continue
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(downloadCmd)
}
