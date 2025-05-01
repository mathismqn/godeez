package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/mathismqn/godeez/internal/downloader"
	"github.com/spf13/cobra"
)

var opts downloader.Options

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download songs from Deezer",
}

func init() {
	RootCmd.AddCommand(downloadCmd)

	downloadCmd.PersistentFlags().StringVarP(&opts.OutputDir, "output", "o", "", "output directory (default is $HOME/Music/GoDeez)")
	downloadCmd.PersistentFlags().StringVarP(&opts.Quality, "quality", "q", "", "download quality [mp3_128, mp3_320, flac, best] (default is best)")

	downloadCmd.AddCommand(
		newDownloadCmd("album"),
		newDownloadCmd("playlist"),
	)
}

func newDownloadCmd(resourceType string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [%s_id...]", resourceType, resourceType),
		Short: fmt.Sprintf("Download songs from one or more %ss", resourceType),
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(appCtx.AppDir)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			dl := downloader.New(appCtx, resourceType)

			if err := dl.Run(ctx, opts, args); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return nil
				}

				return err
			}

			return nil
		},
	}

	return cmd
}
