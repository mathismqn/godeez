package watcher

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureAutostart installs the watcher as a system autostart service.
// Currently disabled: installAutostart is not called due to DB concurrency issues.
func EnsureAutostart(homeDir string) error {
	if isAutostartInstalled(homeDir) || isTemporaryExecutable() {
		return nil
	}

	// return installAutostart(homeDir)

	return nil
}

func isAutostartInstalled(homeDir string) bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	path := filepath.Join(homeDir, "Library", "LaunchAgents", "com.godeez.watch.plist")
	_, err := os.Stat(path)
	return err == nil
}

func isTemporaryExecutable() bool {
	exe, err := os.Executable()
	if err != nil {
		return true
	}
	return strings.Contains(exe, "go-build")
}
