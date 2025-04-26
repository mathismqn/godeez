package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/flytam/filenamify"
	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/db"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/tags"
	"github.com/mathismqn/godeez/internal/utils"
	"github.com/spf13/cobra"
)

var outputDir string
var quality string

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download songs from Deezer",
}

func init() {
	RootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "output directory (default is $HOME/Music/GoDeez)")
	downloadCmd.PersistentFlags().StringVarP(&quality, "quality", "q", "", "download quality [mp3_128, mp3_320, flac, best] (default is best)")
}

func validateInput() {
	if outputDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not get home directory: %v\n", err)
			os.Exit(1)
		}

		musicDir := path.Join(homeDir, "Music")
		if err := utils.EnsureDir(musicDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not create music directory: %v\n", err)
			os.Exit(1)
		}

		outputDir = path.Join(musicDir, "GoDeez")
		if err := utils.EnsureDir(outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not create GoDeez directory: %v\n", err)
			os.Exit(1)
		}
	}

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
		fmt.Fprintf(os.Stderr, "Error: invalid quality option: %s\n", quality)
		os.Exit(1)
	}
}

func downloadContent(contentType string, args []string) {
	session, err := deezer.Authenticate(config.Cfg.ArlCookie)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not authenticate: %v\n", err)
		os.Exit(1)
	}

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
			if err := session.GetData(album, id); err != nil {
				fmt.Printf("\r[%d/%d] Getting data for album %s... FAILED\n", i+1, nArgs, id)
				fmt.Fprintf(os.Stderr, "Error: could not get album data: %v\n", err)

				continue
			}
			resource = album
			songs = album.GetSongs()
		case "playlist":
			playlist := &deezer.Playlist{}
			if err := session.GetData(playlist, id); err != nil {
				fmt.Printf("\r[%d/%d] Getting data for playlist %s... FAILED\n", i+1, nArgs, id)
				fmt.Fprintf(os.Stderr, "Error: could not get playlist data: %v\n", err)

				continue
			}
			if playlist.Results.Data.Status == 1 && playlist.Results.Data.CollabKey == "" {
				fmt.Printf("\r[%d/%d] Getting data for playlist %s... FAILED\n", i+1, nArgs, id)
				fmt.Fprintf(os.Stderr, "Error: playlist is private and no valid arl cookie was provided\n")

				continue
			}

			resource = playlist
			songs = playlist.GetSongs()
		}

		fmt.Printf("\r[%d/%d] Getting data for %s %s... DONE\n", i+1, nArgs, contentType, id)

		output := resource.GetOutputPath(outputDir)
		if err := utils.EnsureDir(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not create output directory: %v\n", err)

			continue
		}

		title := resource.GetTitle()
		fmt.Printf("Starting download of %s: %s\n", contentType, title)

		for _, song := range songs {
			songTitle := song.Title
			if song.Version != "" {
				songTitle = fmt.Sprintf("%s %s", song.Title, song.Version)
			}

			media, err := song.GetMediaData(session.LicenseToken, quality)
			if err != nil {
				fmt.Printf("    Downloading %s... FAILED\n", songTitle)
				fmt.Fprintf(os.Stderr, "Error: could not get media data: %v\n", err)
				if err.Error() == "invalid license token" {
					os.Exit(1)
				}

				continue
			}

			if len(media.Data) == 0 || len(media.Data[0].Media) == 0 || len(media.Data[0].Media[0].Sources) == 0 {
				fmt.Printf("    Downloading %s... FAILED\n", songTitle)
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

			fileName := fmt.Sprintf("%s %s - %s.%s", trackNumber, songTitle, strings.Join(song.Contributors.MainArtists, ", "), ext)
			fileName, _ = filenamify.Filenamify(fileName, filenamify.Options{})
			filePath := path.Join(output, fileName)

			if existing, err := db.Get(song.ID); err == nil {
				if existing.Quality == media.Data[0].Media[0].Format && utils.FileExists(existing.Path) {
					fmt.Printf("    Skipping %s (already downloaded at %s)\n", songTitle, existing.Path)
					continue
				}
			}

			fmt.Printf("    Downloading %s...", songTitle)

			err = media.Download(url, filePath, song.ID)
			if err != nil {
				fmt.Printf("\r    Downloading %s... FAILED\n", songTitle)
				fmt.Fprintf(os.Stderr, "Error: could not download song: %v\n", err)
				utils.DeleteFile(filePath)

				continue
			}
			fmt.Printf("\r    Downloading %s... DONE\n", songTitle)

			tempo, key, err := song.GetTempoAndKey()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not get tempo and key: %v\n", err)
			} else {
				fmt.Printf("        Tempo: %s\n", tempo)
				fmt.Printf("        Key: %s\n", key)
			}

			if err := tags.AddTags(resource, song, filePath, tempo, key); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not add tags to song: %v\n", err)
			}

			info := &db.DownloadInfo{
				SongID:     song.ID,
				Quality:    media.Data[0].Media[0].Format,
				Path:       filePath,
				Downloaded: time.Now(),
			}
			if err := info.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not save download info: %v\n", err)
			}
		}
	}

	fmt.Println(separator)
	fmt.Println("All downloads completed")
}
