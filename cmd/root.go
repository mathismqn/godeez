package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "godeez",
	Short: "GoDeez is a tool to download music from Deezer",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "could not execute command: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.godeez)")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		homedir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get home directory: %v\n", err)
			os.Exit(1)
		}

		path := path.Join(homedir, ".godeez")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("config file not found, creating one at %s\n", path)

			content := []byte("license_token: \nsecret_key: \niv: \n")
			if err := os.WriteFile(path, content, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "could not create config file: %v\n", err)
				os.Exit(1)
			}
		}

		viper.AddConfigPath(homedir)
		viper.SetConfigName(".godeez")
	}

	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "could not read config file: %v\n", err)
		os.Exit(1)
	}

	cfg := &config.Cfg
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "could not unmarshal config: %v\n", err)
		os.Exit(1)
	}

	if cfg.LicenseToken == "" {
		fmt.Fprintln(os.Stderr, "license_token is not set in config file")
		os.Exit(1)
	}
	if cfg.SecretKey == "" {
		fmt.Fprintln(os.Stderr, "secret_key is not set in config file")
		os.Exit(1)
	}
	if cfg.IV == "" {
		fmt.Fprintln(os.Stderr, "iv is not set in config file")
		os.Exit(1)
	}
}
