package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

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

		separator := "--------------------------------------------------"
		for i, id := range args {
			fmt.Println(separator)
			fmt.Printf("[%d/%d] Getting data for album %s...", i+1, nAlbums, id)

			album, err := deezer.GetAlbumData(id)
			if err != nil {
				fmt.Printf("\r[%d/%d] Getting data for album %s... FAILED\n", i+1, nAlbums, id)
				fmt.Fprintf(os.Stderr, "Error: could not get album data: %v\n", err)
				continue
			}

			fmt.Printf("\r[%d/%d] Getting data for album %s... DONE\n", i+1, nAlbums, id)

			output = path.Join(output, fmt.Sprintf("%s - %s", album.Data.Artist, album.Data.Name))
			if _, err := os.Stat(output); os.IsNotExist(err) {
				if err := os.MkdirAll(output, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error: could not create output directory: %v\n", err)
					continue
				}
			}

			fmt.Printf("Starting download of album: %s\n", album.Data.Name)

			for _, song := range album.Songs.Data {
				songTitle := song.Title
				if song.Version != "" {
					songTitle = fmt.Sprintf("%s %s", song.Title, song.Version)
				}

				fmt.Printf("    Downloading %s...", songTitle)

				media, err := song.GetMediaData(quality)
				if err != nil {
					fmt.Printf("\r    Downloading %s... FAILED\n", songTitle)
					fmt.Fprintf(os.Stderr, "Error: could not get media data: %v\n", err)
					if err.Error() == "invalid license token" {
						os.Exit(1)
					}
					continue
				}

				if len(media.Data) == 0 || len(media.Data[0].Media) == 0 || len(media.Data[0].Media[0].Sources) == 0 {
					fmt.Printf("\r    Downloading %s... FAILED\n", songTitle)
					fmt.Fprintf(os.Stderr, "Error: could not get media sources\n")
					continue
				}

				url := media.Data[0].Media[0].Sources[0].URL
				for _, source := range media.Data[0].Media[0].Sources {
					if source.Provider == "ak" {
						url = source.URL
						break
					}
				}
				ext := "mp3"
				if media.Data[0].Media[0].Format == "FLAC" {
					ext = "flac"
				}

				filePath := path.Join(output, fmt.Sprintf("%s - %s.%s", strings.Join(song.Contributors.MainArtists, ", "), songTitle, ext))
				err = media.Download(url, filePath, song.ID)
				if err != nil {
					fmt.Printf("\r    Downloading %s... FAILED\n", songTitle)
					fmt.Fprintf(os.Stderr, "Error: could not download song: %v\n", err)
					continue
				}
				fmt.Printf("\r    Downloading %s... DONE\n", songTitle)

				if err := deezer.AddTags(album, song, filePath); err != nil {
					fmt.Fprintf(os.Stderr, "Error: could not add tags to song: %v\n", err)
				}
			}
		}

		fmt.Println(separator)
		fmt.Println("All downloads completed")
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringP("output", "o", "", "output directory (default is current directory)")
	downloadCmd.Flags().StringP("quality", "q", "", "download quality [mp3_128, mp3_320, flac, best] (default is best)")
}
