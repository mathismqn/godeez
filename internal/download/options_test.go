package download

import (
	"testing"
	"time"

	"github.com/mathismqn/godeez/internal/deezer"
)

func TestOptionsSourceQuality(t *testing.T) {
	tests := []struct {
		quality string
		want    string
		wantWAV bool
	}{
		{"mp3_128", "mp3_128", false},
		{"mp3_320", "mp3_320", false},
		{"flac", "flac", false},
		{"wav", "flac", true},
	}

	for _, tt := range tests {
		t.Run(tt.quality, func(t *testing.T) {
			opts := Options{Quality: tt.quality}

			if got := opts.sourceQuality(); got != tt.want {
				t.Errorf("sourceQuality() = %q, want %q", got, tt.want)
			}
			if got := opts.convertsToWAV(); got != tt.wantWAV {
				t.Errorf("convertsToWAV() = %v, want %v", got, tt.wantWAV)
			}
		})
	}
}

func TestOptionsValidate(t *testing.T) {
	valid := Options{Quality: "mp3_320", Timeout: time.Minute, Limit: 10}

	tests := []struct {
		name    string
		mutate  func(o *Options)
		kind    deezer.Kind
		wantErr bool
	}{
		{"valid", func(o *Options) {}, deezer.KindAlbum, false},
		{"mp3_128", func(o *Options) { o.Quality = "mp3_128" }, deezer.KindAlbum, false},
		{"flac", func(o *Options) { o.Quality = "flac" }, deezer.KindAlbum, false},
		{"wav", func(o *Options) { o.Quality = "wav" }, deezer.KindAlbum, false},
		{"invalid quality", func(o *Options) { o.Quality = "ogg" }, deezer.KindAlbum, true},
		{"uppercase quality", func(o *Options) { o.Quality = "MP3_320" }, deezer.KindAlbum, true},
		{"zero timeout", func(o *Options) { o.Timeout = 0 }, deezer.KindAlbum, true},
		{"negative timeout", func(o *Options) { o.Timeout = -time.Second }, deezer.KindAlbum, true},
		{"artist zero limit", func(o *Options) { o.Limit = 0 }, deezer.KindArtist, true},
		{"artist limit too high", func(o *Options) { o.Limit = 101 }, deezer.KindArtist, true},
		{"artist limit at max", func(o *Options) { o.Limit = 100 }, deezer.KindArtist, false},
		{"album ignores zero limit", func(o *Options) { o.Limit = 0 }, deezer.KindAlbum, false},
		{"track ignores zero limit", func(o *Options) { o.Limit = 0 }, deezer.KindTrack, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := valid
			tt.mutate(&opts)

			err := opts.Validate(tt.kind)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%s) error = %v, wantErr %v", tt.kind, err, tt.wantErr)
			}
		})
	}
}
