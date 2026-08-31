package download

import "testing"

func TestNewDiscLayout(t *testing.T) {
	tests := []struct {
		name       string
		discs      []string
		wantCounts map[string]int
	}{
		{
			name:       "single disc",
			discs:      []string{"1", "1", "1"},
			wantCounts: map[string]int{"1": 3},
		},
		{
			name:       "two discs",
			discs:      []string{"1", "1", "2", "2", "2"},
			wantCounts: map[string]int{"1": 2, "2": 3},
		},
		{
			name:       "missing disc number counts as the first disc",
			discs:      []string{"", "", ""},
			wantCounts: map[string]int{"1": 3},
		},
		{
			name:       "missing disc number joins an explicit first disc",
			discs:      []string{"1", "", "2"},
			wantCounts: map[string]int{"1": 2, "2": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := discLayoutOf(tt.discs...)

			if got := layout.discTotal(); got != len(tt.wantCounts) {
				t.Errorf("discTotal() = %d, want %d", got, len(tt.wantCounts))
			}
			if got, want := layout.multiDisc(), len(tt.wantCounts) > 1; got != want {
				t.Errorf("multiDisc() = %v, want %v", got, want)
			}
			for disc, want := range tt.wantCounts {
				if got := layout.trackTotal(disc); got != want {
					t.Errorf("trackTotal(%q) = %d, want %d", disc, got, want)
				}
			}
		})
	}
}
