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
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		quality, _ := cmd.Flags().GetString("quality")
		nAlbums := len(args)

		if quality == "" {
			quality = "best"
		}
		validQualities := map[string]bool{
			"mp3_128": true,
			"mp3_320": true,
			"flac":    true,
			"best":    true,
		}
		if !validQualities[quality] {
			fmt.Fprintf(os.Stderr, "invalid quality option: %s\n", quality)
			os.Exit(1)
		}

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
				media, err := song.GetMediaData(quality)
				if err != nil {
					fmt.Fprintf(os.Stderr, " could not get media data: %v\n", err)
					if err.Error() == "invalid license token" {
						os.Exit(1)
					}
					continue
				}

				if len(media.Data) == 0 || len(media.Data[0].Media) == 0 || len(media.Data[0].Media[0].Sources) == 0 {
					fmt.Fprintf(os.Stderr, " could not get media sources\n")
					continue
				}

				url := media.Data[0].Media[0].Sources[0].URL
				for _, source := range media.Data[0].Media[0].Sources {
					if source.Provider == "ak" {
						url = source.URL
						break
					}
				}
				songTitle := song.Title
				if song.Version != "" {
					songTitle = fmt.Sprintf("%s %s", song.Title, song.Version)
				}
				ext := "mp3"
				if media.Data[0].Media[0].Format == "FLAC" {
					ext = "flac"
				}

				path := path.Join(output, fmt.Sprintf("%s - %s.%s", song.ArtistName, songTitle, ext))
				err = media.Download(url, path, song.ID)
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
	downloadCmd.Flags().StringP("quality", "q", "", "download quality [mp3_128, mp3_320, flac, best] (default is best)")
}
