package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/db"
	"github.com/mathismqn/godeez/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cfgDir   string
	musicDir string
	appDir   string
	cfgFile  string
)

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
		initDirs()
		config.Init(cfgFile, cfgDir)
		db.Init(cfgDir)
	})
}

func initDirs() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not get home directory: %v\n", err)
		os.Exit(1)
	}

	cfgDir = filepath.Join(home, ".godeez")
	if err := utils.EnsureDir(cfgDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create app directory: %v\n", err)
		os.Exit(1)
	}

	musicDir = filepath.Join(home, "Music")
	if err := utils.EnsureDir(musicDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create music directory: %v\n", err)
		os.Exit(1)
	}

	appDir = path.Join(musicDir, "GoDeez")
	if err := utils.EnsureDir(appDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create GoDeez directory: %v\n", err)
		os.Exit(1)
	}
}
