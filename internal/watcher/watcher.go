package watcher

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/downloader"
	"github.com/mathismqn/godeez/internal/logger"
	"github.com/mathismqn/godeez/internal/store"
)

type Watcher struct {
	appConfig *config.Config
	logger    *logger.Logger
}

func New(appConfig *config.Config) *Watcher {
	homeDir, _ := os.UserHomeDir()
	logFile := filepath.Join(homeDir, ".godeez", "watcher.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v\n", err)
	}

	base := log.New(file, "", log.LstdFlags)
	log := logger.New(base)

	return &Watcher{
		appConfig: appConfig,
		logger:    log,
	}
}

func (w *Watcher) Run(ctx context.Context, opts downloader.Options) {
	w.logger.Infof("Starting watcher...")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			playlists, err := store.ListWatchedPlaylists()
			if err != nil {
				w.logger.Errorf("Failed to list watched playlists: %v\n", err)
			} else {
				for _, playlist := range playlists {
					dl := downloader.New(w.appConfig, "playlist")
					dl.Logger = w.logger
					if err := dl.Run(ctx, opts, playlist.ID); err != nil {
						if errors.Is(err, context.Canceled) {
							return
						}

						w.logger.Errorf("Playlist %s: %v\n", playlist.ID, err)
					}
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(15 * time.Minute):
				// Continue to the next iteration to check for updates
			}
		}
	}
}
