package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/tags"
	"github.com/spf13/cobra"
)

var outputDir string
var quality string

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download songs from Deezer",
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "output directory (default is current directory)")
	downloadCmd.PersistentFlags().StringVarP(&quality, "quality", "q", "", "download quality [mp3_128, mp3_320, flac, best] (default is best)")
}

func validateInput() {
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
}

func downloadContent(contentType string, args []string) {
	nArgs := len(args)
	separator := "--------------------------------------------------"

	for i, id := range args {
		fmt.Println(separator)
		fmt.Printf("[%d/%d] Getting data for %s %s...", i+1, nArgs, contentType, id)

		var resource deezer.Resource
		var songs []*deezer.Song

		switch contentType {
		case "album":
			album := &deezer.Album{}
			if err := deezer.GetData(album, id); err != nil {
				fmt.Printf("\r[%d/%d] Getting data for album %s... FAILED\n", i+1, nArgs, id)
				fmt.Fprintf(os.Stderr, "Error: could not get album data: %v\n", err)
				continue
			}
			resource = album
			songs = album.GetSongs()
		case "playlist":
			playlist := &deezer.Playlist{}
			if err := deezer.GetData(playlist, id); err != nil {
				fmt.Printf("\r[%d/%d] Getting data for playlist %s... FAILED\n", i+1, nArgs, id)
				fmt.Fprintf(os.Stderr, "Error: could not get playlist data: %v\n", err)
				continue
			}
			resource = playlist
			songs = playlist.GetSongs()
		}

		fmt.Printf("\r[%d/%d] Getting data for %s %s... DONE\n", i+1, nArgs, contentType, id)

		output := resource.GetOutputPath(outputDir)
		if _, err := os.Stat(output); os.IsNotExist(err) {
			if err := os.MkdirAll(output, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not create output directory: %v\n", err)
				continue
			}
		}

		title := resource.GetTitle()
		fmt.Printf("Starting download of %s: %s\n", contentType, title)

		for _, song := range songs {
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
			trackNumber := ""
			if contentType == "album" {
				trackNumber = song.TrackNumber + "."
			}

			filePath := path.Join(output, fmt.Sprintf("%s %s - %s.%s", trackNumber, songTitle, strings.Join(song.Contributors.MainArtists, ", "), ext))
			err = media.Download(url, filePath, song.ID)
			if err != nil {
				fmt.Printf("\r    Downloading %s... FAILED\n", songTitle)
				fmt.Fprintf(os.Stderr, "Error: could not download song: %v\n", err)
				continue
			}
			fmt.Printf("\r    Downloading %s... DONE\n", songTitle)

			if err := tags.AddTags(resource, song, filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not add tags to song: %v\n", err)
			}
		}
	}

	fmt.Println(separator)
	fmt.Println("All downloads completed")
}
