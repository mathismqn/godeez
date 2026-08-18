package deezer

import (
	"reflect"
	"testing"
)

func TestNormalizeISRC(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"already normalized", "DE1FB2200042", "DE1FB2200042"},
		{"lowercase", "de1fb2200042", "DE1FB2200042"},
		{"hyphenated", "DE-1FB-22-00042", "DE1FB2200042"},
		{"spaced", "DE 1FB 22 00042", "DE1FB2200042"},
		{"empty", "", ""},
		{"separators only", "- -", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeISRC(tt.in); got != tt.want {
				t.Errorf("normalizeISRC(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSameRecording(t *testing.T) {
	original := &Track{ID: "2358247065", ISRC: "DE1FB2200042", Duration: "164", TrackToken: "original-token"}

	tests := []struct {
		name      string
		candidate *Track
		want      bool
	}{
		{
			name:      "the duplicate Deezer publishes",
			candidate: &Track{ID: "2134121047", ISRC: "DE1FB2200042", Duration: "164", TrackToken: "token"},
			want:      true,
		},
		{
			name:      "duration off by one is a rounding difference",
			candidate: &Track{ID: "2134121047", ISRC: "DE1FB2200042", Duration: "163", TrackToken: "token"},
			want:      true,
		},
		{
			name:      "isrc punctuated differently",
			candidate: &Track{ID: "2134121047", ISRC: "de-1fb-22-00042", Duration: "164", TrackToken: "token"},
			want:      true,
		},
		{
			name:      "duration off by two is a different edit",
			candidate: &Track{ID: "2134121047", ISRC: "DE1FB2200042", Duration: "166", TrackToken: "token"},
			want:      false,
		},
		{
			name:      "different isrc",
			candidate: &Track{ID: "2134121047", ISRC: "DE1FB2300002", Duration: "164", TrackToken: "token"},
			want:      false,
		},
		{
			name:      "candidate has no isrc",
			candidate: &Track{ID: "2134121047", Duration: "164", TrackToken: "token"},
			want:      false,
		},
		{
			name:      "same id",
			candidate: &Track{ID: "2358247065", ISRC: "DE1FB2200042", Duration: "164", TrackToken: "token"},
			want:      false,
		},
		{
			name:      "no id",
			candidate: &Track{ISRC: "DE1FB2200042", Duration: "164", TrackToken: "token"},
			want:      false,
		},
		{
			name:      "no track token",
			candidate: &Track{ID: "2134121047", ISRC: "DE1FB2200042", Duration: "164"},
			want:      false,
		},
		{
			name:      "unparseable duration",
			candidate: &Track{ID: "2134121047", ISRC: "DE1FB2200042", Duration: "unknown", TrackToken: "token"},
			want:      false,
		},
		{
			name:      "zero duration",
			candidate: &Track{ID: "2134121047", ISRC: "DE1FB2200042", Duration: "0", TrackToken: "token"},
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
			if got := sameRecording(original, tt.candidate); got != tt.want {
				t.Errorf("sameRecording() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("original has no isrc", func(t *testing.T) {
		noISRC := &Track{ID: "1", Duration: "164"}
		candidate := &Track{ID: "2", Duration: "164", TrackToken: "token"}
		if sameRecording(noISRC, candidate) {
			t.Error("sameRecording() = true, want false when the original has no ISRC")
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
			seen := map[string]bool{tt.track.ID: true}

			var got []string
			for _, candidate := range embeddedCandidates(tt.track, maxFallbackDepth, seen) {
				got = append(got, candidate.ID)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("embeddedCandidates() = %v, want %v", got, tt.want)
			}

			for _, id := range got {
				if !seen[id] {
					t.Errorf("embeddedCandidates() did not record %q as seen", id)
				}
			}
		})
	}
}

func TestParseISRCLookup(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "readable duplicate",
			body: `{"id":2134121047,"readable":true,"isrc":"DE1FB2200042","duration":164}`,
			want: "2134121047",
		},
		{
			name: "not readable",
			body: `{"id":2134121047,"readable":false,"isrc":"DE1FB2200042"}`,
		},
		{
			name: "no such isrc",
			body: `{"error":{"type":"DataException","message":"no data","code":800}}`,
		},
		{
			name: "quota exceeded",
			body: `{"error":{"type":"Exception","message":"Quota limit exceeded","code":4}}`,
		},
		{
			name: "zero id",
			body: `{"id":0,"readable":true}`,
		},
		{
			name: "malformed body",
			body: `<html>not json</html>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseISRCLookup([]byte(tt.body)); got != tt.want {
				t.Errorf("parseISRCLookup() = %q, want %q", got, tt.want)
			}
		})
	}
}
