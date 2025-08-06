package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mathismqn/godeez/internal/store"
	"github.com/spf13/cobra"
)

var watchAddCmd = &cobra.Command{
	Use:   "add <playlist_id>",
	Short: "Add a playlist to the watch list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		ok, err := store.IsWatched(id)
		if err != nil {
			return err
		}
		if ok {
			fmt.Printf("Playlist %s is already being watched\n", id)
			return nil
		}

		playlist := &store.WatchedPlaylist{
			ID:      id,
			Quality: strings.ToLower(opts.Quality),
			BPM:     opts.BPM,
			Timeout: opts.Timeout,
		}

		if err := playlist.Save(); err != nil {
			return fmt.Errorf("failed to add playlist %s to watch list: %w", id, err)
		}
		fmt.Printf("Playlist %s added to watch list\n", id)

		return nil
	},
}

func init() {
	watchCmd.AddCommand(watchAddCmd)

	watchAddCmd.Flags().StringVarP(&opts.Quality, "quality", "q", "flac", "download quality [mp3_128, mp3_320, flac]")
	watchAddCmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", 2*time.Minute, "timeout for each download (e.g. 10s, 1m, 2m30s)")
	watchAddCmd.Flags().BoolVar(&opts.BPM, "bpm", false, "fetch BPM/key and add to file tags")
}
