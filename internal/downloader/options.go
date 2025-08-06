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
	BPM     bool
	Limit   int
}

func (o *Options) Validate() error {
	if !validQualities[o.Quality] {
		return fmt.Errorf("invalid quality option: %s", o.Quality)
	}
	if o.Limit < 0 {
		return fmt.Errorf("limit must be a non-negative integer")
	}

	return nil
}
