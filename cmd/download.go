package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/downloader"
	"github.com/mathismqn/godeez/internal/store"
	"github.com/spf13/cobra"
)

var opts downloader.Options

var downloadCmd = &cobra.Command{
	Use:         "download",
	Short:       "Download tracks from Deezer",
	Annotations: map[string]string{updateNoticeAnnotation: "true"},
}

func init() {
	RootCmd.AddCommand(downloadCmd)

	downloadCmd.PersistentFlags().StringVarP(&opts.Quality, "quality", "q", "mp3_320", "download quality [mp3_128, mp3_320, flac]")
	downloadCmd.PersistentFlags().DurationVarP(&opts.Timeout, "timeout", "t", 2*time.Minute, "timeout for each download (e.g. 10s, 1m, 2m30s)")
	downloadCmd.PersistentFlags().BoolVar(&opts.BPM, "bpm", false, "fetch BPM/key and add to file tags")
	downloadCmd.PersistentFlags().BoolVar(&opts.Genre, "genre", false, "fetch genre and add to file tags")
	downloadCmd.PersistentFlags().BoolVar(&opts.Strict, "strict", false, "fail the download if the requested quality is unavailable")

	downloadCmd.AddCommand(
		newDownloadCmd(deezer.KindAlbum),
		newDownloadCmd(deezer.KindPlaylist),
		newDownloadCmd(deezer.KindArtist),
		newDownloadCmd(deezer.KindTrack),
	)
}

func newDownloadCmd(kind deezer.Kind) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <%s_id>", kind, kind),
		Short: downloadShort(kind),
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opts.Quality = strings.ToLower(opts.Quality)
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			config.MigrateLegacy(cfg.OutputDir)

			st, err := store.Open(cfg.OutputDir)
			if err != nil {
				return err
			}
			defer st.Close()

			err = downloader.New(cfg, st, kind).Run(cmd.Context(), opts, args[0])
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		},
	}

	if kind == deezer.KindArtist {
		cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 10, "number of tracks to download")
	}

	return cmd
}

func downloadShort(kind deezer.Kind) string {
	switch kind {
	case deezer.KindArtist:
		return "Download an artist's top tracks"
	case deezer.KindTrack:
		return "Download a single track"
	case deezer.KindAlbum:
		return "Download tracks from an album"
	default:
		return fmt.Sprintf("Download tracks from a %s", kind)
	}
}
