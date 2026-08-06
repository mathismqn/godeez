package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/download"
	"github.com/mathismqn/godeez/internal/store"
	"github.com/spf13/cobra"
)

func newDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "download",
		Short:       "Download tracks from Deezer",
		Annotations: map[string]string{updateNoticeAnnotation: "true"},
	}

	opts := &download.Options{}
	cmd.PersistentFlags().StringVarP(&opts.Quality, "quality", "q", "mp3_320", "download quality [mp3_128, mp3_320, flac, wav]")
	cmd.PersistentFlags().DurationVarP(&opts.Timeout, "timeout", "t", 2*time.Minute, "timeout for each download (e.g. 10s, 1m, 2m30s)")
	cmd.PersistentFlags().BoolVar(&opts.BPM, "bpm", false, "fetch BPM/key and add to file tags")
	cmd.PersistentFlags().BoolVar(&opts.Genre, "genre", false, "fetch genre and add to file tags")
	cmd.PersistentFlags().BoolVar(&opts.Strict, "strict", false, "fail the download if the requested quality is unavailable")

	cmd.AddCommand(
		newDownloadSubCmd(deezer.KindAlbum, opts),
		newDownloadSubCmd(deezer.KindPlaylist, opts),
		newDownloadSubCmd(deezer.KindArtist, opts),
		newDownloadSubCmd(deezer.KindTrack, opts),
	)

	return cmd
}

// newDownloadSubCmd builds one download subcommand from a deezer.Kind. The
// four kinds differ only in wording and in whether they take a track limit,
// so they share this constructor rather than being written out four times.
//
// All four share one Options value through the parent's persistent flags,
// which is safe because exactly one subcommand ever runs.
//
// A cancelled download is reported as success: the user pressed Ctrl-C and
// has already seen the progress output, so an error on top of it would be
// noise, and a non-zero exit would misreport a deliberate stop as a failure.
func newDownloadSubCmd(kind deezer.Kind, opts *download.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <%s_id>", kind, kind),
		Short: downloadShort(kind),
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opts.Quality = strings.ToLower(opts.Quality)
			return opts.Validate(kind)
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

			err = download.New(cfg, st, kind).Run(cmd.Context(), *opts, args[0])
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
