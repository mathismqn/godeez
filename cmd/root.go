package cmd

import (
	"fmt"
	"os"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/watcher"
	"github.com/spf13/cobra"
)

var (
	cfgPath   string
	appConfig *config.Config
)

var RootCmd = &cobra.Command{
	Use:          "godeez",
	Short:        "GoDeez is a tool to download music from Deezer",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		appConfig, err = config.New(cfgPath)
		if err != nil {
			return err
		}

		if err := watcher.EnsureAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to install autostart for watcher: %v\n", err)
		}

		return nil
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file (default ~/.godeez/config.toml)")
}
