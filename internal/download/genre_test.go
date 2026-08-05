package download

import (
	"slices"
	"testing"
)

func TestMatchesKeyword(t *testing.T) {
	if !matchesKeyword("Deep House", electronicKeywords) {
		t.Error("expected Deep House to match electronic keywords")
	}
	if !matchesKeyword("classic rock", nonElectronicKeywords) {
		t.Error("expected classic rock to match non-electronic keywords")
	}
	if matchesKeyword("Spoken Word", electronicKeywords) {
		t.Error("did not expect Spoken Word to match electronic keywords")
	}
}

func TestFilterTags(t *testing.T) {
	got := filterTags([]string{"Deep House", "Rock", "Spoken Word"})
	want := []string{"Deep House", "Rock"}
	if !slices.Equal(got, want) {
		t.Errorf("filterTags() = %v, want %v", got, want)
	}

	if got := filterTags([]string{"Rock", "Jazz"}); len(got) != 0 {
		t.Errorf("filterTags() = %v, want empty", got)
	}

	if got := filterTags(nil); len(got) != 0 {
		t.Errorf("filterTags(nil) = %v, want empty", got)
	}
}

func TestFormatTags(t *testing.T) {
	tests := []struct {
		tags []string
		want string
	}{
		{[]string{"deep house"}, "Deep House"},
		{[]string{"TECHNO", "trance"}, "Techno / Trance"},
		{[]string{" house ", ""}, "House"},
	}

	for _, tt := range tests {
		if got := formatTags(tt.tags); got != tt.want {
			t.Errorf("formatTags(%v) = %q, want %q", tt.tags, got, tt.want)
		}
	}
}
