package download

import (
	"testing"
	"time"
)

func TestOptionsValidate(t *testing.T) {
	valid := Options{Quality: "mp3_320", Timeout: time.Minute, Limit: 10}

	tests := []struct {
		name    string
		mutate  func(o *Options)
		wantErr bool
	}{
		{"valid", func(o *Options) {}, false},
		{"mp3_128", func(o *Options) { o.Quality = "mp3_128" }, false},
		{"flac", func(o *Options) { o.Quality = "flac" }, false},
		{"invalid quality", func(o *Options) { o.Quality = "ogg" }, true},
		{"uppercase quality", func(o *Options) { o.Quality = "MP3_320" }, true},
		{"zero timeout", func(o *Options) { o.Timeout = 0 }, true},
		{"negative timeout", func(o *Options) { o.Timeout = -time.Second }, true},
		{"zero limit", func(o *Options) { o.Limit = 0 }, true},
		{"limit too high", func(o *Options) { o.Limit = 101 }, true},
		{"limit at max", func(o *Options) { o.Limit = 100 }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := valid
			tt.mutate(&opts)

			err := opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
