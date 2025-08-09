package downloader

import (
	"fmt"
	"time"
)

var validQualities = map[string]bool{
	"mp3_128": true,
	"mp3_320": true,
	"flac":    true,
}

type Options struct {
	Quality string
	Timeout time.Duration
	Limit   int
	BPM     bool
	Strict  bool
}

func (o *Options) Validate() error {
	if !validQualities[o.Quality] {
		return fmt.Errorf("invalid quality option: %s", o.Quality)
	}
	if o.Timeout <= 0 {
		return fmt.Errorf("timeout must be a positive duration")
	}
	if o.Limit <= 0 {
		return fmt.Errorf("limit must be a positive integer")
	}
	if o.Limit > 100 {
		return fmt.Errorf("limit must not exceed 100")
	}

	return nil
}
