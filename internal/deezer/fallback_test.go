package deezer

import (
	"reflect"
	"testing"
)

// variantOf returns a copy of base with edits applied, so a table case states
// only the fields it varies.
func variantOf(base Track, edits ...func(*Track)) *Track {
	candidate := base
	for _, edit := range edits {
		edit(&candidate)
	}

	return &candidate
}

func TestLinkedStandIn(t *testing.T) {
	original := &Track{ID: "1984831597", Artist: "cults", Title: "Gilded Lily", ISRC: "QM8QH1700573", Duration: "213", TrackToken: "original-token"}

	// base is a valid stand-in for original.
	base := Track{ID: "2075423167", Artist: "cults", Title: "Gilded Lily", ISRC: "USQX92206420", Duration: "212", TrackToken: "token"}
	with := func(edits ...func(*Track)) *Track { return variantOf(base, edits...) }

	tests := []struct {
		name      string
		candidate *Track
		want      bool
	}{
		{
			name:      "the re-release Deezer links to, under a new isrc",
			candidate: with(),
			want:      true,
		},
		{
			name:      "isrc matching as well",
			candidate: with(func(c *Track) { c.ISRC = original.ISRC }),
			want:      true,
		},
		{
			name:      "candidate carries no isrc",
			candidate: with(func(c *Track) { c.ISRC = "" }),
			want:      true,
		},
		{
			name:      "artist and title cased differently",
			candidate: with(func(c *Track) { c.Artist = "Cults"; c.Title = "GILDED LILY" }),
			want:      true,
		},
		{
			name:      "duration off by two is a trimmed fade",
			candidate: with(func(c *Track) { c.Duration = "215" }),
			want:      true,
		},
		{
			name:      "duration off by five is the edge of the window",
			candidate: with(func(c *Track) { c.Duration = "208" }),
			want:      true,
		},
		{
			name:      "duration off by six is a different edit",
			candidate: with(func(c *Track) { c.Duration = "219" }),
			want:      false,
		},
		{
			name:      "a radio edit of the same song",
			candidate: with(func(c *Track) { c.Duration = "178" }),
			want:      false,
		},
		{
			name:      "a live take of the same song",
			candidate: with(func(c *Track) { c.Version = "(Live)" }),
			want:      false,
		},
		{
			name:      "another song by the same artist",
			candidate: with(func(c *Track) { c.Title = "Always Forever" }),
			want:      false,
		},
		{
			name:      "a cover by someone else",
			candidate: with(func(c *Track) { c.Artist = "Someone Else" }),
			want:      false,
		},
		{
			name:      "unparseable duration",
			candidate: with(func(c *Track) { c.Duration = "unknown" }),
			want:      false,
		},
		{
			name:      "same id",
			candidate: with(func(c *Track) { c.ID = original.ID }),
			want:      false,
		},
		{
			name:      "no track token",
			candidate: with(func(c *Track) { c.TrackToken = "" }),
			want:      false,
		},
		{
			name:      "empty fallback object",
			candidate: &Track{},
			want:      false,
		},
		{
			name:      "no candidate",
			candidate: nil,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := linkedStandIn(original, tt.candidate); got != tt.want {
				t.Errorf("linkedStandIn() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("original has no duration", func(t *testing.T) {
		noDuration := &Track{ID: "1", Artist: "cults", Title: "Gilded Lily"}
		candidate := &Track{ID: "2", Artist: "cults", Title: "Gilded Lily", Duration: "213", TrackToken: "token"}
		if linkedStandIn(noDuration, candidate) {
			t.Error("linkedStandIn() = true, want false when the original has no duration")
		}
	})
}

func TestEmbeddedCandidates(t *testing.T) {
	chain := func(ids ...string) *Track {
		var head *Track
		for i := len(ids) - 1; i >= 0; i-- {
			head = &Track{ID: ids[i], Fallback: head}
		}
		return head
	}

	selfCycle := &Track{ID: "a"}
	selfCycle.Fallback = selfCycle

	cycle := &Track{ID: "a"}
	cycle.Fallback = &Track{ID: "b", Fallback: cycle}

	tests := []struct {
		name  string
		track *Track
		want  []string
	}{
		{"no fallback", &Track{ID: "a"}, nil},
		{"one level", chain("a", "b"), []string{"b"}},
		{"chain longer than the depth limit", chain("a", "b", "c", "d", "e"), []string{"b", "c", "d"}},
		{"self cycle", selfCycle, nil},
		{"cycle back to the original", cycle, []string{"b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, candidate := range embeddedCandidates(tt.track, maxFallbackDepth) {
				got = append(got, candidate.ID)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("embeddedCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoverCandidates(t *testing.T) {
	const (
		ownCover     = "b940289c9ccb1981f85876cf83311efd"
		standInCover = "c95730cdd2eada45468be38317067b8e"
		furtherCover = "3b1e0a1f7c2d4e5a6b7c8d9e0f1a2b3c"
		unrelatedArt = "ffffffffffffffffffffffffffffffff"
	)

	// original is the dead entry, base a valid stand-in for it.
	original := Track{ID: "2967949521", Artist: "Gaskin", Title: "Closer", Duration: "234", TrackToken: "original-token"}
	base := Track{ID: "3811506992", Artist: "Gaskin", Title: "Closer", Duration: "234", TrackToken: "token", Cover: standInCover}
	with := func(edits ...func(*Track)) *Track { return variantOf(base, edits...) }
	from := func(cover string, fallback *Track) *Track {
		return variantOf(original, func(t *Track) { t.Cover = cover; t.Fallback = fallback })
	}

	tests := []struct {
		name  string
		track *Track
		want  []string
	}{
		{
			name:  "the track has its own cover",
			track: from(ownCover, with()),
			want:  []string{ownCover, standInCover},
		},
		{
			name:  "no cover of its own, the stand-in has one",
			track: from("", with()),
			want:  []string{standInCover},
		},
		{
			name:  "no cover anywhere",
			track: from("", with(func(c *Track) { c.Cover = "" })),
			want:  nil,
		},
		{
			name:  "the placeholder id spelled out is not a cover",
			track: from(missingCoverMD5, with()),
			want:  []string{standInCover},
		},
		{
			name:  "the placeholder id on both is no cover at all",
			track: from(missingCoverMD5, with(func(c *Track) { c.Cover = missingCoverMD5 })),
			want:  nil,
		},
		{
			name:  "no fallback at all",
			track: from("", nil),
			want:  nil,
		},
		{
			name:  "a candidate that is not a stand-in is not asked for its sleeve",
			track: from("", with(func(c *Track) { c.Duration = "178"; c.Cover = unrelatedArt })),
			want:  nil,
		},
		{
			name:  "a candidate with no track token is not a stand-in",
			track: from("", with(func(c *Track) { c.TrackToken = "" })),
			want:  nil,
		},
		{
			name: "further down the chain",
			track: from("", with(func(c *Track) {
				c.Cover = ""
				c.Fallback = with(func(f *Track) { f.ID = "4000000001"; f.Cover = furtherCover })
			})),
			want: []string{furtherCover},
		},
		{
			name: "the same cover twice in the chain",
			track: from(standInCover, with(func(c *Track) {
				c.Fallback = with(func(f *Track) { f.ID = "4000000001" })
			})),
			want: []string{standInCover},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := coverCandidates(tt.track); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("coverCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}
