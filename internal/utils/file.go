package utils

import "os"

func FileExists(path string) bool {
	info, err := os.Stat(path)

	return err == nil && !info.IsDir()
}
