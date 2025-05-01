package cmd

import (
	"github.com/mathismqn/godeez/internal/app"
	"github.com/spf13/cobra"
)

var (
	cfgPath string
	appCtx  *app.Context
)

var RootCmd = &cobra.Command{
	Use:          "godeez",
	Short:        "GoDeez is a tool to download music from Deezer",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		appCtx, err = app.NewContext(cfgPath)

		return err
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file (default is $HOME/.godeez)")
}
