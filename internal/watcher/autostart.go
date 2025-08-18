package watcher

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func EnsureAutostart(homeDir string) error {
	if isAutostartInstalled(homeDir) || isTemporaryExecutable() {
		return nil
	}

	// return installAutostart(homeDir)

	return nil
}

func isAutostartInstalled(homeDir string) bool {
	switch runtime.GOOS {
	case "darwin":
		path := filepath.Join(homeDir, "Library", "LaunchAgents", "com.godeez.watch.plist")
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
