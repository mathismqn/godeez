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
		nAlbums := len(args)

		for i, id := range args {
			fmt.Printf("[%d/%d] Getting data for album %s...", i+1, nAlbums, id)
			album, err := deezer.GetAlbumData(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\ncould not get album data: %v\n", err)
				continue
			}
			fmt.Println(" done")
			fmt.Printf("Starting download of %s", album.Data.Name)

			for _, song := range album.Songs.Data {
				fmt.Printf("\nDownloading %s...", song.Title)
				media, err := song.GetMediaData()
				if err != nil {
					fmt.Fprintf(os.Stderr, " could not get media data: %v\n", err)
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
					fmt.Fprintf(os.Stderr, " could not download song: %v\n", err)
					continue
				}
				fmt.Print(" done")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
