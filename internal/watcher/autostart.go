package watcher

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func EnsureAutostart() error {
	if isAutostartInstalled() || isTemporaryExecutable() {
		return nil
	}

	return installAutostart()
}

func isAutostartInstalled() bool {
	switch runtime.GOOS {
	case "darwin":
		path := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.godeez.watch.plist")
		_, err := os.Stat(path)

		return err == nil
	default:
		return false
	}
}

func isTemporaryExecutable() bool {
	exe, err := os.Executable()
	if err != nil {
		return true
	}

	return strings.Contains(exe, "go-build")
}
