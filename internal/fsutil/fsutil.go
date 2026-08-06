// Package fsutil holds the few filesystem helpers shared across godeez.
package fsutil

import (
	"fmt"
	"os"
)

// PartPattern names in-progress downloads. It is both an os.CreateTemp
// pattern and a glob, which is what lets the downloader sweep away leftovers
// from an interrupted run. The leading dot hides them from file managers, and
// the suffix keeps them from being mistaken for finished audio.
const PartPattern = ".godeez-*.part"

// EnsureDir creates path if it does not exist. An existing non-directory at
// that path is an error rather than something to overwrite.
func EnsureDir(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("file already exists at %s", path)
	}
	return nil
}

// Exists reports whether path is an existing regular file. A directory is
// deliberately not "exists" here: every caller is asking about a file it
// intends to read, write or delete.
func Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func Remove(path string) error {
	if !Exists(path) {
		return nil
	}
	return os.Remove(path)
}
