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

func New(cfgPath, cfgDir string) (*Config, error) {
	if cfgPath != "" {
		viper.SetConfigFile(cfgPath)
	} else {
		cfgPath := path.Join(cfgDir, "config.toml")
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			fmt.Printf("Config file not found, creating one at %s\n", cfgPath)

			content := []byte("arl_cookie = ''\nsecret_key = ''\niv = '0001020304050607'\n")
			if err := os.WriteFile(cfgPath, content, 0644); err != nil {
				return nil, fmt.Errorf("failed to create config file: %w", err)
			}
		}

		viper.AddConfigPath(cfgDir)
		viper.SetConfigName("config.toml")
	}

	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.ArlCookie == "" {
		return nil, fmt.Errorf("arl_cookie is not set in config file")
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("secret_key is not set in config file")
	}
	if len(cfg.SecretKey) != 16 {
		return nil, fmt.Errorf("secret_key must be 16 bytes long")
	}
	if cfg.IV == "" {
		return nil, fmt.Errorf("iv is not set in config file")
	}
	if len(cfg.IV) != 16 {
		return nil, fmt.Errorf("iv must be 16 bytes long")
	}

	return cfg, nil
}
