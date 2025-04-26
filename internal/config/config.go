package config

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

type Config struct {
	ArlCookie string `mapstructure:"arl_cookie"`
	SecretKey string `mapstructure:"secret_key"`
	IV        string `mapstructure:"iv"`
}

var Cfg Config

func Init(cfgFile, cfgDir string) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		cfgPath := path.Join(cfgDir, "config.toml")
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			fmt.Printf("Config file not found, creating one at %s\n", cfgPath)

			content := []byte("arl_cookie = ''\nsecret_key = ''\niv = '0001020304050607'\n")
			if err := os.WriteFile(cfgPath, content, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not create config file: %v\n", err)
				os.Exit(1)
			}
		}

		viper.AddConfigPath(cfgDir)
		viper.SetConfigName("config.toml")
	}

	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read config file: %v\n", err)
		os.Exit(1)
	}

	cfg := &Cfg
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not unmarshal config file: %v\n", err)
		os.Exit(1)
	}

	if cfg.SecretKey == "" {
		fmt.Fprintln(os.Stderr, "Error: secret_key is not set in config file")
		os.Exit(1)
	}
	if cfg.IV == "" {
		fmt.Fprintln(os.Stderr, "Error: iv is not set in config file")
		os.Exit(1)
	}
}
