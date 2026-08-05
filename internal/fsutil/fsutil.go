package fsutil

import (
	"fmt"
	"os"
)

const PartPattern = ".godeez-*.part"

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
