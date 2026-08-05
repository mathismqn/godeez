package config

import (
	"os"
	"path/filepath"
)

func MigrateLegacy(outputDir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	oldDB := filepath.Join(homeDir, ".godeez", "tracks.db")
	newDB := filepath.Join(outputDir, ".tracks.db")
	if _, err := os.Stat(oldDB); err == nil {
		if _, err := os.Stat(newDB); os.IsNotExist(err) {
			os.Rename(oldDB, newDB)
		}
	}

	oldDir := filepath.Join(homeDir, ".godeez")
	os.Remove(filepath.Join(oldDir, "config.toml"))
	os.Remove(oldDir)
}
