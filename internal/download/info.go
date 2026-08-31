package download

import (
	"fmt"
	"strings"
	"time"

	"github.com/mathismqn/godeez/internal/deezer"
)

func resourceInfo(resource deezer.Resource) string {
	switch r := resource.(type) {
	case *deezer.Album:
		return albumInfo(r)
	case *deezer.Playlist:
		return playlistInfo(r)
	case *deezer.Artist:
		return artistInfo(r)
	case *deezer.Single:
		return singleInfo(r)
	}

	return ""
}

// playtime converts a gateway duration in seconds. These banners are
// cosmetic, so a field Deezer omitted counts as zero rather than as an error.
func playtime(seconds deezer.Number) time.Duration {
	n, ok := seconds.Int()
	if !ok {
		return 0
	}

	return time.Duration(n) * time.Second
}

func albumInfo(a *deezer.Album) string {
	return fmt.Sprintf(
		`================= [ Album Info ] =================
Title:    %s
Artist:   %s
Tracks:   %d
Duration: %s
==================================================`,
		a.Results.Data.Title,
		a.Results.Data.Artist,
		len(a.Results.Tracks.Data),
		playtime(a.Results.Data.Duration),
	)
}

func playlistInfo(p *deezer.Playlist) string {
	return fmt.Sprintf(
		`=============== [ Playlist Info ] ===============
Title:    %s
Creator:  %s
Tracks:   %d
Duration: %s
=================================================`,
		p.Results.Data.Title,
		p.Results.Data.Creator,
		len(p.Results.Tracks.Data),
		playtime(p.Results.Data.Duration),
	)
}

func artistInfo(a *deezer.Artist) string {
	tracks := a.Results.Tracks.Data
	count := len(tracks)

	var total time.Duration
	for _, t := range tracks {
		total += playtime(t.Duration)
	}

	limit := min(3, count)

	var b strings.Builder
	fmt.Fprintf(&b, "============= [ Artist Info ] =============\n")
	fmt.Fprintf(&b, "Artist:   %s\n", a.Results.Data.Name)
	fmt.Fprintf(&b, "Tracks:   %d\n", count)
	fmt.Fprintf(&b, "Playtime: %s\n", total)
	fmt.Fprintf(&b, "-------------------------------------------\n")
	fmt.Fprintf(&b, "Top %d most popular tracks:\n", limit)
	for i := 0; i < limit; i++ {
		t := tracks[i]
		fmt.Fprintf(&b, "    %2d. %s – %s\n", i+1, t.Artist, t.FullTitle())
	}
	fmt.Fprintf(&b, "===========================================\n")

	return b.String()
}

func singleInfo(s *deezer.Single) string {
	if s.Results.Data == nil {
		return "Track: No data available"
	}

	return fmt.Sprintf(
		`================= [ Track Info ] =================
Title:    %s
Artist:   %s
Duration: %s
==================================================`,
		s.Results.Data.FullTitle(),
		s.Results.Data.Artist,
		playtime(s.Results.Data.Duration),
	)
}
