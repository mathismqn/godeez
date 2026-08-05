package tag

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func minimalWAV(audio []byte) []byte {
	var body bytes.Buffer
	body.WriteString("WAVE")

	fmtPayload := make([]byte, 0, 16)
	fmtPayload = binary.LittleEndian.AppendUint16(fmtPayload, 1)
	fmtPayload = binary.LittleEndian.AppendUint16(fmtPayload, 2)
	fmtPayload = binary.LittleEndian.AppendUint32(fmtPayload, 44100)
	fmtPayload = binary.LittleEndian.AppendUint32(fmtPayload, 176400)
	fmtPayload = binary.LittleEndian.AppendUint16(fmtPayload, 4)
	fmtPayload = binary.LittleEndian.AppendUint16(fmtPayload, 16)
	writeChunk(&body, "fmt ", fmtPayload)
	writeChunk(&body, "data", audio)

	var out bytes.Buffer
	out.WriteString("RIFF")
	binary.Write(&out, binary.LittleEndian, uint32(body.Len()))
	out.Write(body.Bytes())

	return out.Bytes()
}

func parseChunks(t *testing.T, data []byte) map[string][]byte {
	t.Helper()

	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		t.Fatalf("bad riff header: %q %q", data[0:4], data[8:12])
	}
	if size := binary.LittleEndian.Uint32(data[4:8]); int(size) != len(data)-8 {
		t.Errorf("riff size = %d, want %d", size, len(data)-8)
	}

	chunks := make(map[string][]byte)
	for offset := 12; offset < len(data); {
		if offset%2 != 0 {
			t.Errorf("chunk at offset %d is not word aligned", offset)
		}
		if offset+8 > len(data) {
			t.Fatalf("truncated chunk header at offset %d", offset)
		}

		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		if offset+8+size > len(data) {
			t.Fatalf("chunk %q at offset %d overruns the file", id, offset)
		}
		if _, ok := chunks[id]; ok {
			t.Errorf("duplicate %q chunk", id)
		}
		chunks[id] = data[offset+8 : offset+8+size]

		offset += 8 + size + size%2
	}

	return chunks
}

func testMetadata() Metadata {
	return Metadata{
		Title:       "Song",
		Artists:     "Artist",
		Genre:       "Rock",
		BPM:         "120",
		Key:         "Am",
		TrackNumber: "3",
		Duration:    "215",
		Gain:        "-7.5",
		ISRC:        "FR1234567890",
		Cover:       []byte{0xff, 0xd8, 0xff, 0xe0, 0x00},
		Album: &AlbumMetadata{
			Artist:       "Album Artist",
			Title:        "Album",
			Label:        "Label",
			ReleaseDate:  "2024-05-01",
			ProducerLine: "Producer line",
			Copyright:    "Copyright",
		},
	}
}

func writeTestWAV(t *testing.T, audio []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "track.wav")
	if err := os.WriteFile(path, minimalWAV(audio), 0644); err != nil {
		t.Fatalf("write wav: %v", err)
	}

	return path
}

func TestWriteWAV(t *testing.T) {
	audio := bytes.Repeat([]byte{0x11, 0x22, 0x33, 0x44}, 16)
	path := writeTestWAV(t, audio)
	original := parseChunks(t, minimalWAV(audio))

	if err := Write(path, testMetadata()); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read wav: %v", err)
	}
	chunks := parseChunks(t, data)

	if !bytes.Equal(chunks["data"], original["data"]) {
		t.Error("data chunk was modified")
	}
	if !bytes.Equal(chunks["fmt "], original["fmt "]) {
		t.Error("fmt chunk was modified")
	}

	id3, ok := chunks["id3 "]
	if !ok {
		t.Fatal("missing id3 chunk")
	}
	if string(id3[0:3]) != "ID3" {
		t.Errorf("id3 chunk does not start with an ID3 header: %q", id3[0:3])
	}
	for _, want := range []string{"Song", "Artist", "Album", "120", "Am", "FR1234567890", "Rock"} {
		if !bytes.Contains(id3, []byte(want)) {
			t.Errorf("id3 chunk is missing %q", want)
		}
	}
	if !bytes.Contains(id3, []byte{0xff, 0xd8, 0xff, 0xe0}) {
		t.Error("id3 chunk is missing the cover art")
	}

	list, ok := chunks["LIST"]
	if !ok {
		t.Fatal("missing LIST chunk")
	}
	if string(list[0:4]) != "INFO" {
		t.Errorf("LIST form = %q, want \"INFO\"", list[0:4])
	}
	for _, want := range []struct{ id, value string }{
		{"INAM", "Song"},
		{"IART", "Artist"},
		{"IGNR", "Rock"},
		{"ITRK", "3"},
		{"IPRD", "Album"},
		{"ICRD", "2024"},
		{"ICMT", "Producer line"},
		{"ICOP", "Copyright"},
	} {
		if !bytes.Contains(list, append([]byte(want.id), append([]byte{byte(len(want.value) + 1), 0, 0, 0}, want.value...)...)) {
			t.Errorf("LIST chunk is missing %s = %q", want.id, want.value)
		}
	}
}

func TestWriteWAVIsIdempotent(t *testing.T) {
	audio := bytes.Repeat([]byte{0x01, 0x02}, 32)
	path := writeTestWAV(t, audio)

	if err := Write(path, testMetadata()); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read wav: %v", err)
	}
	firstChunks := parseChunks(t, first)

	if err := Write(path, testMetadata()); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read wav: %v", err)
	}
	secondChunks := parseChunks(t, second)

	if len(first) != len(second) {
		t.Errorf("re-tagging changed the file size: %d bytes then %d bytes", len(first), len(second))
	}
	if len(firstChunks) != len(secondChunks) {
		t.Errorf("chunk count = %d, want %d", len(secondChunks), len(firstChunks))
	}
	for id, payload := range firstChunks {
		got, ok := secondChunks[id]
		if !ok {
			t.Errorf("re-tagging dropped the %q chunk", id)
			continue
		}
		if len(got) != len(payload) {
			t.Errorf("%q chunk size = %d, want %d", id, len(got), len(payload))
		}
	}
	if !bytes.Equal(secondChunks["data"], audio) {
		t.Error("data chunk was modified")
	}
	if !bytes.Equal(secondChunks["fmt "], firstChunks["fmt "]) {
		t.Error("fmt chunk was modified")
	}
	if !bytes.Equal(secondChunks["LIST"], firstChunks["LIST"]) {
		t.Error("LIST chunk was modified")
	}
}

func TestWriteWAVOddSizedChunks(t *testing.T) {
	audio := bytes.Repeat([]byte{0x07}, 33)
	path := writeTestWAV(t, audio)

	if err := Write(path, testMetadata()); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read wav: %v", err)
	}

	chunks := parseChunks(t, data)
	if !bytes.Equal(chunks["data"], audio) {
		t.Error("data chunk was modified")
	}
	if _, ok := chunks["id3 "]; !ok {
		t.Error("missing id3 chunk")
	}
}

func TestWriteWAVRejectsNonWAV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "track.wav")
	if err := os.WriteFile(path, []byte("this is not a wav file at all"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := Write(path, testMetadata()); err == nil {
		t.Error("Write() error = nil, want error")
	}
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Error("Write() left a temp file behind")
	}
}
