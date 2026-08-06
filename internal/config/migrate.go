package config

import (
	"io"
	"os"
	"path/filepath"
)

// MigrateLegacy moves the download ledger from the old ~/.godeez directory
// next to the user's music, where it now lives.
//
// It is silent and best effort throughout. A failed migration costs the user
// their skip history, which the next download simply rebuilds, so there is
// nothing worth interrupting them about. An existing database at the new
// location always wins, which makes this safe to run on every download rather
// than needing a flag to say whether it has happened yet.
//
// The rename is attempted first and falls back to a copy, because the old and
// new locations are often on different filesystems.
func MigrateLegacy(outputDir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	oldDir := filepath.Join(homeDir, ".godeez")
	oldDB := filepath.Join(oldDir, "tracks.db")
	newDB := filepath.Join(outputDir, ".tracks.db")
	if _, err := os.Stat(oldDB); err == nil {
		if _, err := os.Stat(newDB); os.IsNotExist(err) {
			if err := os.Rename(oldDB, newDB); err != nil {
				if err := copyFile(oldDB, newDB); err != nil {
					return
				}
				os.Remove(oldDB)
			}
		}
	}

	os.Remove(oldDir)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.CreateTemp(filepath.Dir(dst), ".tracks.db-*.tmp")
	if err != nil {
		return err
	}
	tmp := out.Name()

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Sync(); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dst); err != nil {
		os.Remove(tmp)
		return err
	}

	return nil
}
