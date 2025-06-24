//go:build darwin

package watcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func installAutostart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
 "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.godeez.watch</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>watch</string>
    <string>run</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
</dict>
</plist>`, exe)

	path := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "com.godeez.watch.plist")
	if err := os.WriteFile(path, []byte(plist), 0644); err != nil {
		return err
	}

	return exec.Command("launchctl", "load", path).Run()
}
