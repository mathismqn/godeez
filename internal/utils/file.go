package utils

import "os"

func FileExists(path string) bool {
	info, err := os.Stat(path)

	return err == nil && !info.IsDir()
}

func DeleteFile(path string) error {
	if !FileExists(path) {
		return nil
	}

	return os.Remove(path)
}
