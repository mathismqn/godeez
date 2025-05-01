package app

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/fileutil"
	"github.com/mathismqn/godeez/internal/store"
)

type Context struct {
	AppDir string
	Config *config.Config
}

func NewContext(cfgPath string) (*Context, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cfgDir := filepath.Join(home, ".godeez")
	if err := fileutil.EnsureDir(cfgDir); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	musicDir := filepath.Join(home, "Music")
	if err := fileutil.EnsureDir(musicDir); err != nil {
		return nil, fmt.Errorf("failed to create music directory: %w", err)
	}

	appDir := path.Join(musicDir, "GoDeez")
	if err := fileutil.EnsureDir(appDir); err != nil {
		return nil, fmt.Errorf("failed to create app directory: %w", err)
	}

	cfg, err := config.New(cfgPath, cfgDir)
	if err != nil {
		return nil, err
	}

	if err := store.OpenDB(cfgDir); err != nil {
		return nil, err
	}

	return &Context{
		AppDir: appDir,
		Config: cfg,
	}, nil
}
