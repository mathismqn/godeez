package download

import (
	"errors"
	"fmt"
	"time"

	"github.com/mathismqn/godeez/internal/deezer"
)

var validQualities = map[string]bool{
	"mp3_128": true,
	"mp3_320": true,
	"flac":    true,
	"wav":     true,
}

type Options struct {
	Quality string
	Timeout time.Duration
	Limit   int
	BPM     bool
	Genre   bool
	Strict  bool
}

func (o *Options) sourceQuality() string {
	if o.Quality == "wav" {
		return "flac"
	}
	return o.Quality
}

func (o *Options) convertsToWAV() bool {
	return o.Quality == "wav"
}

func (o *Options) Validate(kind deezer.Kind) error {
	if !validQualities[o.Quality] {
		return fmt.Errorf("invalid quality option: %s", o.Quality)
	}
	if o.Timeout <= 0 {
		return errors.New("timeout must be a positive duration")
	}
	if kind == deezer.KindArtist {
		if o.Limit <= 0 {
			return errors.New("limit must be a positive integer")
		}
		if o.Limit > 100 {
			return errors.New("limit must not exceed 100")
		}
	}

	return nil
}
