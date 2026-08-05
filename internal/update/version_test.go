package update

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"1.0.0", "1.0.1", true},
		{"v1.0.0", "v1.1.0", true},
		{"1.0.0", "v2.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.1.0", "1.0.0", false},
		{" 1.0.0 ", "1.0.1", true},
		{"1.0.0", "1.0.1-rc.1", true},
		{"1.0.0-rc.1", "1.0.0", true},
		{"dev", "1.0.0", false},
		{"1.0.0", "not-a-version", false},
		{"", "1.0.0", false},
		{"1.0.0", "", false},
	}

	for _, tt := range tests {
		if got := IsNewer(tt.current, tt.latest); got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestTrimV(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"v1.2.3", "1.2.3"},
		{"1.2.3", "1.2.3"},
		{" v1.2.3 ", "1.2.3"},
	}

	for _, tt := range tests {
		if got := trimV(tt.in); got != tt.want {
			t.Errorf("trimV(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCanonical(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"1.2.3", "v1.2.3"},
		{"v1.2.3", "v1.2.3"},
		{"", ""},
		{"garbage", ""},
	}

	for _, tt := range tests {
		if got := canonical(tt.in); got != tt.want {
			t.Errorf("canonical(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
