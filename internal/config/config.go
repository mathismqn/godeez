// Package config resolves where godeez reads its session from and writes its
// downloads to. There is no config file: the output directory is fixed and
// the only setting is the DEEZER_ARL environment variable, which exists as an
// escape hatch for users who would rather not store credentials in the system
// keyring.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mathismqn/godeez/internal/fsutil"
)

type Config struct {
	ARLCookie string
	OutputDir string
}

// Load resolves the configuration and creates the output directory.
//
// An empty ARLCookie is normal and not an error: it means fall back to the
// stored credentials, which is the usual path. Creating the directory here
// rather than at first write means a bad path fails immediately instead of
// after the first track has been fetched.
func Load() (*Config, error) {
	arl := os.Getenv("DEEZER_ARL")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	outputDir := filepath.Join(homeDir, "Music", "GoDeez")
	if err := fsutil.EnsureDir(outputDir); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &Config{
		ARLCookie: arl,
		OutputDir: outputDir,
	}, nil
}
