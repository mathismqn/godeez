package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/mathismqn/godeez/internal/store"
	"github.com/spf13/cobra"
)

var watchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List watched playlists",
	RunE: func(cmd *cobra.Command, args []string) error {
		playlists, err := store.ListWatchedPlaylists()
		if err != nil {
			return fmt.Errorf("failed to list watched playlists: %w", err)
		}

		if len(playlists) == 0 {
			fmt.Println("No watched playlists.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tQuality\tFetch BPM\tTimeout")
		fmt.Fprintln(w, "---\t-------\t----------\t-------")

		for _, playlist := range playlists {
			fmt.Fprintf(w, "%s\t%s\t%t\t%s\n", playlist.ID, playlist.Quality, playlist.BPM, playlist.Timeout)
		}
		w.Flush()

		return nil
	},
}

func init() {
	watchCmd.AddCommand(watchListCmd)
}
