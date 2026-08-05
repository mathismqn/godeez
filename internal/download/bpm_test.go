package download

import "testing"

func bpmHTML(bpm, key, mode string) string {
	return `tempo of <span class="x">` + bpm + ` BPM</span>` +
		` with a <span class="x">` + key + `</span> key` +
		` and a  <span class="x">` + mode + `</span> mode`
}

func TestParseBPM(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantBPM string
		wantKey string
	}{
		{"major key", bpmHTML("128", "A", "major"), "128", "A"},
		{"minor key gets m suffix", bpmHTML("90", "F", "minor"), "90", "Fm"},
		{"unicode sharp normalized", bpmHTML("124", "C♯", "major"), "124", "C#"},
		{"unicode flat normalized", bpmHTML("100", "B♭", "minor"), "100", "Bbm"},
		{"enharmonic pair keeps first", bpmHTML("110", "A♯/B♭", "major"), "110", "A#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBPM(tt.html)
			if err != nil {
				t.Fatalf("parseBPM: %v", err)
			}
			if got.BPM != tt.wantBPM || got.Key != tt.wantKey {
				t.Errorf("parseBPM() = %+v, want BPM %q Key %q", got, tt.wantBPM, tt.wantKey)
			}
		})
	}
}

func TestParseBPMNoData(t *testing.T) {
	if _, err := parseBPM("<html>nothing here</html>"); err == nil {
		t.Error("expected error for page without BPM data")
	}
}
