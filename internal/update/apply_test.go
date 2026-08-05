package update

import (
	"strings"
	"testing"
)

func TestParseChecksums(t *testing.T) {
	checksums := `ABCDEF0123  godeez_1.0.0_darwin_arm64
deadbeef  *godeez_1.0.0_linux_amd64
malformed-line
one two three
cafebabe  godeez_1.0.0_windows_amd64.exe
`

	tests := []struct {
		name  string
		asset string
		want  string
	}{
		{"plain name lowercased", "godeez_1.0.0_darwin_arm64", "abcdef0123"},
		{"star-prefixed name", "godeez_1.0.0_linux_amd64", "deadbeef"},
		{"windows asset", "godeez_1.0.0_windows_amd64.exe", "cafebabe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseChecksums(strings.NewReader(checksums), tt.asset)
			if err != nil {
				t.Fatalf("parseChecksums: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseChecksums() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseChecksumsMissing(t *testing.T) {
	if _, err := parseChecksums(strings.NewReader("abc  other_asset\n"), "godeez_1.0.0_darwin_arm64"); err == nil {
		t.Error("expected error for missing asset name")
	}
}
