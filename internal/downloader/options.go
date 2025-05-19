package downloader

import (
	"fmt"
	"time"
)

var validQualities = map[string]bool{
	"mp3_128": true,
	"mp3_320": true,
	"flac":    true,
	"best":    true,
}

type Options struct {
	OutputDir string
	Quality   string
	Timeout   time.Duration
	BPM       bool
}

func (o *Options) Validate(appDir string) error {
	if o.OutputDir == "" {
		o.OutputDir = appDir
	}

	if o.Quality == "" {
		o.Quality = "best"
	}

	if !validQualities[o.Quality] {
		return fmt.Errorf("invalid quality option: %s", o.Quality)
	}

	return nil
}
