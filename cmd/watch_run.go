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
		cmd.SetContext(context.WithValue(cmd.Context(), "appConfig", appConfig))
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		appConfigVal := ctx.Value("appConfig")
		appConfig, _ := appConfigVal.(*config.Config)

		w := watcher.New(appConfig)
		w.Run(ctx, opts)
	},
}

func init() {
	watchCmd.AddCommand(watchRunCmd)
}
