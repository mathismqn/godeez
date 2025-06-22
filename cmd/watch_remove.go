package cmd

import (
	"fmt"

	"github.com/mathismqn/godeez/internal/store"
	"github.com/spf13/cobra"
)

var watchRemoveCmd = &cobra.Command{
	Use:   "remove <playlist_id>",
	Short: "Remove a playlist from the watch list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		ok, err := store.IsWatched(id)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("playlist %s is not being watched", id)
		}

		if err := store.RemoveWatchedPlaylist(id); err != nil {
			return fmt.Errorf("failed to remove playlist %s from watch list: %w", id, err)
		}
		fmt.Printf("Playlist %s removed from watch list\n", id)

		return nil
	},
}

func init() {
	watchCmd.AddCommand(watchRemoveCmd)
}
