package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/downloader"
	"github.com/spf13/cobra"
)

type contextKey string

const appConfigKey contextKey = "appConfig"

var opts downloader.Options

var downloadCmd = &cobra.Command{
	Use:         "download",
	Short:       "Download songs from Deezer",
	Annotations: map[string]string{updateNoticeAnnotation: "true"},
}

func init() {
	RootCmd.AddCommand(downloadCmd)

	downloadCmd.PersistentFlags().StringVarP(&opts.Quality, "quality", "q", "mp3_320", "download quality [mp3_128, mp3_320, flac]")
	downloadCmd.PersistentFlags().DurationVarP(&opts.Timeout, "timeout", "t", 2*time.Minute, "timeout for each download (e.g. 10s, 1m, 2m30s)")
	downloadCmd.PersistentFlags().BoolVar(&opts.BPM, "bpm", false, "fetch BPM/key and add to file tags")
	downloadCmd.PersistentFlags().BoolVar(&opts.Genre, "genre", false, "fetch genre and add to file tags")
	downloadCmd.PersistentFlags().BoolVar(&opts.Strict, "strict", false, "fail the song download if the quality is not available")

	downloadCmd.AddCommand(
		newDownloadCmd("album"),
		newDownloadCmd("playlist"),
		newDownloadCmd("artist"),
		newDownloadCmd("track"),
	)
}

func newDownloadCmd(resourceType string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <%s_id>", resourceType, resourceType),
		Short: downloadShort(resourceType),
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			appConfig, err := config.New()
			if err != nil {
				return err
			}
			cmd.SetContext(context.WithValue(cmd.Context(), appConfigKey, appConfig))

			opts.Quality = strings.ToLower(opts.Quality)
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			appConfig := cmd.Context().Value(appConfigKey).(*config.Config)

			err := downloader.New(appConfig, resourceType).Run(cmd.Context(), opts, args[0])
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		},
	}

	if resourceType == "artist" {
		cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 10, "number of songs to download")
	}

	return cmd
}

func downloadShort(resourceType string) string {
	switch resourceType {
	case "artist":
		return "Download top songs from an artist"
	case "track":
		return "Download a single track"
	case "album":
		return "Download songs from an album"
	default:
		return fmt.Sprintf("Download songs from a %s", resourceType)
	}
}
