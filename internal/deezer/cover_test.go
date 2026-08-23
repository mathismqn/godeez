package deezer

import (
	"net/url"
	"testing"
)

func TestCoverURL(t *testing.T) {
	const cover = "b940289c9ccb1981f85876cf83311efd"

	want := "https://e-cdn-images.dzcdn.net/images/cover/" + cover + "/500x500-000000-80-0-0.jpg"
	if got := coverURL(cover); got != want {
		t.Errorf("coverURL() = %q, want %q", got, want)
	}
}

func TestIsPlaceholderCoverURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{
			name: "where the cdn redirects an empty or unknown id",
			raw:  coverURL(missingCoverMD5),
			want: true,
		},
		{
			name: "the request that gets redirected, before it is followed",
			raw:  coverURL(""),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.raw)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got := isPlaceholderCoverURL(u); got != tt.want {
				t.Errorf("isPlaceholderCoverURL(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}

	t.Run("no url", func(t *testing.T) {
		if isPlaceholderCoverURL(nil) {
			t.Error("isPlaceholderCoverURL(nil) = true, want false")
		}
	})
}
