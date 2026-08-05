package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mathismqn/godeez/internal/fileutil"
)

type Config struct {
	ARLCookie string
	OutputDir string
}

func Load() (*Config, error) {
	arl := os.Getenv("DEEZER_ARL")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	outputDir := filepath.Join(homeDir, "Music", "GoDeez")
	if err := fileutil.EnsureDir(outputDir); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &Config{
		ARLCookie: arl,
		OutputDir: outputDir,
	}, nil
}
