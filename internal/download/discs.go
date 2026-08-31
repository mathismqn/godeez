package download

import "github.com/mathismqn/godeez/internal/deezer"

// discLayout counts the tracks on each disc of an album. The counts come from
// the song list rather than from an album level field, so they hold even when
// Deezer's own totals disagree with the tracks it actually returned.
type discLayout map[string]int

// discNumber is the disc a track belongs to. Deezer omits DISK_NUMBER on some
// releases, and treating that as disc one is what keeps those albums named
// exactly as they always were.
func discNumber(track *deezer.Track) string {
	if track.DiscNumber == "" {
		return "1"
	}

	return string(track.DiscNumber)
}

func newDiscLayout(tracks []*deezer.Track) discLayout {
	layout := make(discLayout)
	for _, track := range tracks {
		layout[discNumber(track)]++
	}

	return layout
}

func (l discLayout) multiDisc() bool {
	return len(l) > 1
}

func (l discLayout) discTotal() int {
	return len(l)
}

func (l discLayout) trackTotal(disc string) int {
	return l[disc]
}
