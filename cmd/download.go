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

var (
	opts    downloader.Options
	cfgPath string
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download songs from Deezer",
}

func init() {
	RootCmd.AddCommand(downloadCmd)

	downloadCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file (default ~/.godeez/config.toml)")
	downloadCmd.PersistentFlags().StringVarP(&opts.Quality, "quality", "q", "best", "download quality [mp3_128, mp3_320, flac, best]")
	downloadCmd.PersistentFlags().DurationVarP(&opts.Timeout, "timeout", "t", 2*time.Minute, "timeout for each download (e.g. 10s, 1m, 2m30s)")
	downloadCmd.PersistentFlags().BoolVar(&opts.BPM, "bpm", false, "fetch BPM/key and add to file tags")

	downloadCmd.AddCommand(
		newDownloadCmd("album"),
		newDownloadCmd("playlist"),
		newDownloadCmd("artist"),
	)
}

func newDownloadCmd(resourceType string) *cobra.Command {
	article := "a"
	if resourceType == "album" {
		article = "an"
	}

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <%s_id>", resourceType, resourceType),
		Short: fmt.Sprintf("Download songs from %s %s", article, resourceType),
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			appConfig, err := config.New(cfgPath)
			if err != nil {
				return err
			}
			cmd.SetContext(context.WithValue(cmd.Context(), "appConfig", appConfig))

			opts.Quality = strings.ToLower(opts.Quality)

			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			appConfigVal := ctx.Value("appConfig")
			appConfig, _ := appConfigVal.(*config.Config)

			dl := downloader.New(appConfig, resourceType)
			if err := dl.Run(ctx, opts, args[0]); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}

				return err
			}

			return nil
		},
	}

	if resourceType == "artist" {
		cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 10, "number of songs to download")
		cmd.Short = fmt.Sprintf("Download top songs from an artist")
	}

	return cmd
}
