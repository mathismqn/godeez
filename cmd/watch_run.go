package cmd

import (
	"github.com/mathismqn/godeez/internal/watcher"
	"github.com/spf13/cobra"
)

var watchRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Start the background playlist watcher",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		w := watcher.New(appConfig)
		w.Run(ctx, opts)
	},
}

func init() {
	watchCmd.AddCommand(watchRunCmd)
}
