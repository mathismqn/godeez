package cmd

import (
	"context"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/watcher"
	"github.com/spf13/cobra"
)

var watchRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Start the background playlist watcher",
	Hidden: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		appConfig, err := config.New("")
		if err != nil {
			return
		}
		cmd.SetContext(context.WithValue(cmd.Context(), appConfigKey, appConfig))
	},
	Run: func(cmd *cobra.Command, args []string) {
		appConfig, _ := cmd.Context().Value(appConfigKey).(*config.Config)
		watcher.New(appConfig).Run(cmd.Context(), opts)
	},
}

func init() {
	watchCmd.AddCommand(watchRunCmd)
}
