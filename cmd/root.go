package cmd

import (
	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/db"
	"github.com/spf13/cobra"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "godeez",
	Short: "GoDeez is a tool to download music from Deezer",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.godeez)")
	cobra.OnInitialize(func() {
		config.Init(cfgFile)
		db.Init()
	})
}
