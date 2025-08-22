package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mathismqn/godeez/internal/fileutil"
	"github.com/mathismqn/godeez/internal/store"
	"github.com/spf13/viper"
)

type Config struct {
	ArlCookie string `mapstructure:"arl_cookie"`
	SecretKey string `mapstructure:"secret_key"`
	OutputDir string `mapstructure:"output_dir"`
	HomeDir   string
	ConfigDir string
}

func New(cfgPath string) (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cfgDir := filepath.Join(homeDir, ".godeez")
	if err := fileutil.EnsureDir(cfgDir); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	if cfgPath == "" {
		cfgPath = path.Join(cfgDir, "config.toml")
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			fmt.Printf("Config file not found, creating one at %s\n", cfgPath)

			content := []byte("arl_cookie = ''\nsecret_key = ''\noutput_dir = ''\n")
			if err := os.WriteFile(cfgPath, content, 0644); err != nil {
				return nil, fmt.Errorf("failed to create config file: %w", err)
			}

			os.Exit(0)
		}
	}

	viper.SetConfigFile(cfgPath)
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{
		HomeDir:   homeDir,
		ConfigDir: cfgDir,
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if err := store.OpenDB(cfgDir); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.ArlCookie == "" {
		return fmt.Errorf("arl_cookie is not set")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("secret_key is not set")
	}
	if len(c.SecretKey) != 16 {
		return fmt.Errorf("secret_key must be 16 bytes long")
	}
	if c.OutputDir == "" {
		c.OutputDir = filepath.Join(c.HomeDir, "Music", "GoDeez")
	}

	return nil
}
