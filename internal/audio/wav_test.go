package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
	"github.com/mewkiz/flac/meta"
)

func testSamples(nChannels, bitsPerSample, nSamples int) [][]int32 {
	max := int32(1)<<(bitsPerSample-1) - 1
	min := -int32(1) << (bitsPerSample - 1)

	channels := make([][]int32, nChannels)
	for c := range channels {
		samples := make([]int32, nSamples)
		for i := range samples {
			switch i {
			case 0:
				samples[i] = min
			case 1:
				samples[i] = max
			case 2:
				samples[i] = 0
			default:
				samples[i] = int32(i*(c+1)) % max
				if i%3 == 0 {
					samples[i] = -samples[i]
				}
			}
		}
		channels[c] = samples
	}

	return channels
}

func writeTestFLAC(t *testing.T, path string, sampleRate uint32, bitsPerSample uint8, channels [][]int32) {
	t.Helper()

	nSamples := len(channels[0])
	info := &meta.StreamInfo{
		BlockSizeMin:  uint16(nSamples),
		BlockSizeMax:  uint16(nSamples),
		SampleRate:    sampleRate,
		NChannels:     uint8(len(channels)),
		BitsPerSample: bitsPerSample,
		NSamples:      uint64(nSamples),
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create flac: %v", err)
	}
	defer file.Close()

	enc, err := flac.NewEncoder(file, info)
	if err != nil {
		t.Fatalf("new encoder: %v", err)
	}

	subframes := make([]*frame.Subframe, len(channels))
	for c, samples := range channels {
		subframes[c] = &frame.Subframe{
			SubHeader: frame.SubHeader{Pred: frame.PredVerbatim},
			Samples:   samples,
			NSamples:  nSamples,
		}
	}

	channelsLayout := frame.ChannelsMono
	if len(channels) == 2 {
		channelsLayout = frame.ChannelsLR
	}

	f := &frame.Frame{
		Header: frame.Header{
			HasFixedBlockSize: true,
			BlockSize:         uint16(nSamples),
			SampleRate:        sampleRate,
			Channels:          channelsLayout,
			BitsPerSample:     bitsPerSample,
		},
		Subframes: subframes,
	}

	if err := enc.WriteFrame(f); err != nil {
		t.Fatalf("write frame: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("close encoder: %v", err)
	}
}

func expectedPCM(channels [][]int32, bytesPerSample int) []byte {
	var buf bytes.Buffer

	for i := range channels[0] {
		for _, samples := range channels {
			sample := samples[i]
			switch bytesPerSample {
			case 1:
				buf.WriteByte(byte(sample + 128))
			case 2:
				buf.Write([]byte{byte(sample), byte(sample >> 8)})
			case 3:
				buf.Write([]byte{byte(sample), byte(sample >> 8), byte(sample >> 16)})
			}
		}
	}

	return buf.Bytes()
}

func TestFLACToWAV(t *testing.T) {
	tests := []struct {
		name          string
		sampleRate    uint32
		bitsPerSample uint8
		nChannels     int
		nSamples      int
	}{
		{"16 bit stereo", 44100, 16, 2, 512},
		{"16 bit mono", 44100, 16, 1, 512},
		{"24 bit stereo", 48000, 24, 2, 333},
		{"8 bit mono", 22050, 8, 1, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "in.flac")
			dst := filepath.Join(dir, "out.wav")

			channels := testSamples(tt.nChannels, int(tt.bitsPerSample), tt.nSamples)
			writeTestFLAC(t, src, tt.sampleRate, tt.bitsPerSample, channels)

			if err := FLACToWAV(context.Background(), src, dst); err != nil {
				t.Fatalf("FLACToWAV() error = %v", err)
			}

			got, err := os.ReadFile(dst)
			if err != nil {
				t.Fatalf("read wav: %v", err)
			}

			bytesPerSample := int(tt.bitsPerSample) / 8
			blockAlign := uint16(tt.nChannels * bytesPerSample)
			want := expectedPCM(channels, bytesPerSample)
			dataSize := uint32(len(want))
			pad := dataSize % 2

			if len(got) != headerSize+len(want)+int(pad) {
				t.Fatalf("file size = %d, want %d", len(got), headerSize+len(want)+int(pad))
			}
			if string(got[0:4]) != "RIFF" || string(got[8:12]) != "WAVE" {
				t.Errorf("magic = %q %q, want \"RIFF\" \"WAVE\"", got[0:4], got[8:12])
			}
			if size := binary.LittleEndian.Uint32(got[4:8]); size != headerSize-8+dataSize+pad {
				t.Errorf("riff size = %d, want %d", size, headerSize-8+dataSize+pad)
			}
			if string(got[12:16]) != "fmt " {
				t.Errorf("fmt chunk id = %q, want \"fmt \"", got[12:16])
			}
			if size := binary.LittleEndian.Uint32(got[16:20]); size != 16 {
				t.Errorf("fmt chunk size = %d, want 16", size)
			}
			if format := binary.LittleEndian.Uint16(got[20:22]); format != formatPCM {
				t.Errorf("format = %d, want %d", format, formatPCM)
			}
			if n := binary.LittleEndian.Uint16(got[22:24]); n != uint16(tt.nChannels) {
				t.Errorf("channels = %d, want %d", n, tt.nChannels)
			}
			if rate := binary.LittleEndian.Uint32(got[24:28]); rate != tt.sampleRate {
				t.Errorf("sample rate = %d, want %d", rate, tt.sampleRate)
			}
			if rate := binary.LittleEndian.Uint32(got[28:32]); rate != tt.sampleRate*uint32(blockAlign) {
				t.Errorf("byte rate = %d, want %d", rate, tt.sampleRate*uint32(blockAlign))
			}
			if align := binary.LittleEndian.Uint16(got[32:34]); align != blockAlign {
				t.Errorf("block align = %d, want %d", align, blockAlign)
			}
			if bits := binary.LittleEndian.Uint16(got[34:36]); bits != uint16(tt.bitsPerSample) {
				t.Errorf("bits per sample = %d, want %d", bits, tt.bitsPerSample)
			}
			if string(got[36:40]) != "data" {
				t.Errorf("data chunk id = %q, want \"data\"", got[36:40])
			}
			if size := binary.LittleEndian.Uint32(got[40:44]); size != dataSize {
				t.Errorf("data chunk size = %d, want %d", size, dataSize)
			}
			if !bytes.Equal(got[headerSize:headerSize+len(want)], want) {
				t.Error("pcm payload does not match the source samples")
			}
		})
	}
}

func TestFLACToWAVErrors(t *testing.T) {
	dir := t.TempDir()

	notFLAC := filepath.Join(dir, "not.flac")
	if err := os.WriteFile(notFLAC, []byte("this is not a flac file"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	valid := filepath.Join(dir, "valid.flac")
	writeTestFLAC(t, valid, 44100, 16, testSamples(2, 16, 64))
	data, err := os.ReadFile(valid)
	if err != nil {
		t.Fatalf("read flac: %v", err)
	}
	truncated := filepath.Join(dir, "truncated.flac")
	if err := os.WriteFile(truncated, data[:len(data)/2], 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{"not a flac file", notFLAC},
		{"missing file", filepath.Join(dir, "missing.flac")},
		{"truncated file", truncated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := filepath.Join(t.TempDir(), "out.wav")
			if err := FLACToWAV(context.Background(), tt.path, dst); err == nil {
				t.Error("FLACToWAV() error = nil, want error")
			}
			if _, err := os.Stat(dst); err == nil {
				t.Error("FLACToWAV() left an output file behind")
			}
		})
	}
}

func TestFLACToWAVCanceled(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.flac")
	dst := filepath.Join(dir, "out.wav")
	writeTestFLAC(t, src, 44100, 16, testSamples(2, 16, 512))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := FLACToWAV(ctx, src, dst); err == nil {
		t.Error("FLACToWAV() error = nil, want context.Canceled")
	}
	if _, err := os.Stat(dst); err == nil {
		t.Error("FLACToWAV() left an output file behind")
	}
	matches, _ := filepath.Glob(filepath.Join(dir, ".godeez-*.part"))
	if len(matches) > 0 {
		t.Errorf("FLACToWAV() left %d part files behind", len(matches))
	}
}

func TestBytesPerSample(t *testing.T) {
	tests := []struct {
		bitsPerSample uint8
		want          int
		wantErr       bool
	}{
		{8, 1, false},
		{16, 2, false},
		{24, 3, false},
		{4, 0, true},
		{12, 0, true},
		{20, 0, true},
		{32, 0, true},
	}

	for _, tt := range tests {
		got, err := bytesPerSample(tt.bitsPerSample)
		if (err != nil) != tt.wantErr {
			t.Errorf("bytesPerSample(%d) error = %v, wantErr %v", tt.bitsPerSample, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("bytesPerSample(%d) = %d, want %d", tt.bitsPerSample, got, tt.want)
		}
	}
}
