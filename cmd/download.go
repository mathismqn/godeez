package cmd

import (
	"fmt"
	"os"
	"path"

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

		output, _ := cmd.Flags().GetString("output")
		nAlbums := len(args)

		for i, id := range args {
			fmt.Printf("[%d/%d] Getting data for album %s...", i+1, nAlbums, id)
			album, err := deezer.GetAlbumData(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\ncould not get album data: %v\n", err)
				continue
			}
			fmt.Println(" done")

			output = path.Join(output, fmt.Sprintf("%s - %s", album.Data.ArtistName, album.Data.Name))
			if _, err := os.Stat(output); os.IsNotExist(err) {
				if err := os.MkdirAll(output, 0755); err != nil {
					fmt.Fprintf(os.Stderr, " could not create output directory: %v\n", err)
					continue
				}
			}

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

				path := path.Join(output, fmt.Sprintf("%s - %s.flac", song.ArtistName, songTitle))
				err = media.Download(path, song.ID)
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
	downloadCmd.Flags().StringP("output", "o", "", "output directory (default is current directory)")
}
