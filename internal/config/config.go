package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mathismqn/godeez/internal/fileutil"
	"github.com/mathismqn/godeez/internal/store"
)

type Config struct {
	ARLCookie string
	OutputDir string
	HomeDir   string
}

func New() (*Config, error) {
	arl := os.Getenv("DEEZER_ARL")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	outputDir := filepath.Join(homeDir, "Music", "GoDeez")
	if err := fileutil.EnsureDir(outputDir); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Migrate tracks.db from ~/.godeez/ to output dir
	oldDB := filepath.Join(homeDir, ".godeez", "tracks.db")
	newDB := filepath.Join(outputDir, ".tracks.db")
	if _, err := os.Stat(oldDB); err == nil {
		if _, err := os.Stat(newDB); os.IsNotExist(err) {
			os.Rename(oldDB, newDB)
		}
	}

	// Clean up old config directory
	oldDir := filepath.Join(homeDir, ".godeez")
	os.Remove(filepath.Join(oldDir, "config.toml"))
	os.Remove(oldDir) // fails silently if not empty

	if err := store.OpenDB(outputDir); err != nil {
		return nil, err
	}

	return &Config{
		ARLCookie: arl,
		OutputDir: outputDir,
		HomeDir:   homeDir,
	}, nil
}
