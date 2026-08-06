// Package audio converts downloaded audio between formats.
//
// Deezer does not serve wav, so a wav download is really a flac download
// followed by FLACToWAV. The conversion is lossless in both directions: flac
// decodes to exactly the PCM samples it was encoded from, so nothing is lost
// by going through it.
package audio

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/mathismqn/godeez/internal/fsutil"
	"github.com/mewkiz/flac"
)

const (
	// headerSize is the canonical PCM wav header: a 12 byte RIFF/WAVE header,
	// a 24 byte fmt chunk, and an 8 byte data chunk header.
	headerSize = 44
	formatPCM  = 1

	// ctxCheckInterval is how often, in flac frames, cancellation is polled.
	// A flac frame is a few thousand samples, so checking every frame would
	// add a select to the innermost decode loop for no practical gain in
	// responsiveness.
	ctxCheckInterval = 64

	// maxDataSize is the largest audio payload that still fits. RIFF stores
	// its sizes as uint32, and the RIFF size field covers the header after
	// its own first 8 bytes as well as the data, so the audio itself has to
	// stay that much below the limit. This works out to roughly 6 hours of
	// CD quality stereo, which no single track will reach, but silently
	// producing a file with a wrapped size field would be worse than an
	// error.
	maxDataSize = math.MaxUint32 - (headerSize - 8)
)

// FLACToWAV decodes the flac at srcPath and writes it as a PCM wav to
// dstPath.
//
// The size is checked twice, once from the flac header before doing any work
// and once against the bytes actually written, because NSamples is zero in
// flac streams that were encoded without a known length.
//
// Output goes to a temporary file in the destination directory and is renamed
// into place at the end, so a cancelled or failed conversion never leaves a
// half decoded file where a playable one is expected.
func FLACToWAV(ctx context.Context, srcPath, dstPath string) error {
	stream, err := flac.Open(srcPath)
	if err != nil {
		return err
	}
	defer stream.Close()

	info := stream.Info
	bytesPerSample, err := bytesPerSample(info.BitsPerSample)
	if err != nil {
		return err
	}
	if info.NChannels < 1 || info.NChannels > 2 {
		return fmt.Errorf("unsupported channel count: %d", info.NChannels)
	}
	if size := int64(info.NSamples) * int64(info.NChannels) * int64(bytesPerSample); size > maxDataSize {
		return fmt.Errorf("audio data of %d bytes exceeds the wav format limit", size)
	}

	file, err := os.CreateTemp(filepath.Dir(dstPath), fsutil.PartPattern)
	if err != nil {
		return err
	}
	tmpPath := file.Name()
	done := false
	defer func() {
		if !done {
			file.Close()
			os.Remove(tmpPath)
		}
	}()

	// The header goes down with a zero data size and is patched afterwards:
	// the real length is only known once every frame has been decoded, and
	// buffering the whole stream in memory to find out first is not worth it.
	w := bufio.NewWriter(file)
	if err := writeHeader(w, info.SampleRate, info.NChannels, info.BitsPerSample, 0); err != nil {
		return err
	}

	dataSize, err := writeSamples(ctx, w, stream, int(info.NChannels), bytesPerSample)
	if err != nil {
		return err
	}
	if dataSize > maxDataSize {
		return fmt.Errorf("audio data of %d bytes exceeds the wav format limit", dataSize)
	}
	// RIFF chunks must end on an even offset. Only reachable with 8 or 24 bit
	// mono, where a sample is an odd number of bytes.
	if dataSize%2 != 0 {
		if err := w.WriteByte(0); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := patchSizes(file, dataSize); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		return err
	}
	done = true

	return nil
}

func bytesPerSample(bitsPerSample uint8) (int, error) {
	switch bitsPerSample {
	case 8, 16, 24:
		return int(bitsPerSample) / 8, nil
	default:
		return 0, fmt.Errorf("unsupported bit depth: %d", bitsPerSample)
	}
}

func writeHeader(w io.Writer, sampleRate uint32, nChannels, bitsPerSample uint8, dataSize uint32) error {
	blockAlign := uint32(nChannels) * uint32(bitsPerSample) / 8

	header := make([]byte, 0, headerSize)
	header = append(header, "RIFF"...)
	header = binary.LittleEndian.AppendUint32(header, uint32(headerSize-8)+dataSize)
	header = append(header, "WAVE"...)
	header = append(header, "fmt "...)
	header = binary.LittleEndian.AppendUint32(header, 16)
	header = binary.LittleEndian.AppendUint16(header, formatPCM)
	header = binary.LittleEndian.AppendUint16(header, uint16(nChannels))
	header = binary.LittleEndian.AppendUint32(header, sampleRate)
	header = binary.LittleEndian.AppendUint32(header, sampleRate*blockAlign)
	header = binary.LittleEndian.AppendUint16(header, uint16(blockAlign))
	header = binary.LittleEndian.AppendUint16(header, uint16(bitsPerSample))
	header = append(header, "data"...)
	header = binary.LittleEndian.AppendUint32(header, dataSize)

	_, err := w.Write(header)

	return err
}

func writeSamples(ctx context.Context, w io.Writer, stream *flac.Stream, nChannels, bytesPerSample int) (int64, error) {
	var dataSize int64
	buf := make([]byte, 4)

	for i := 0; ; i++ {
		if i%ctxCheckInterval == 0 {
			select {
			case <-ctx.Done():
				return dataSize, ctx.Err()
			default:
			}
		}

		frame, err := stream.ParseNext()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return dataSize, err
		}
		if len(frame.Subframes) != nChannels {
			return dataSize, fmt.Errorf("frame %d has %d channels, want %d", frame.Num, len(frame.Subframes), nChannels)
		}

		for i := range frame.Subframes[0].Samples {
			for _, subframe := range frame.Subframes {
				putSample(buf, subframe.Samples[i], bytesPerSample)
				if _, err := w.Write(buf[:bytesPerSample]); err != nil {
					return dataSize, err
				}
				dataSize += int64(bytesPerSample)
			}
		}
	}

	return dataSize, nil
}

// putSample encodes one sample little endian into buf.
//
// 8 bit wav is the odd one out: it stores unsigned samples biased by 128,
// while every wider depth is signed two's complement. Writing an 8 bit sample
// signed produces audio that sounds like loud static, so the bias is not
// optional.
func putSample(buf []byte, sample int32, bytesPerSample int) {
	if bytesPerSample == 1 {
		buf[0] = byte(sample + 128)
		return
	}

	value := uint32(sample)
	for i := range bytesPerSample {
		buf[i] = byte(value >> (8 * i))
	}
}

// patchSizes rewrites the two length fields once the real data size is known:
// the RIFF size at offset 4 and the data chunk size just before the samples
// begin.
//
// The pad byte counts towards the RIFF size but not towards the data chunk
// size, which is why only the first of the two includes it.
func patchSizes(file *os.File, dataSize int64) error {
	buf := make([]byte, 4)

	binary.LittleEndian.PutUint32(buf, uint32(headerSize-8+dataSize+dataSize%2))
	if _, err := file.WriteAt(buf, 4); err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(buf, uint32(dataSize))
	_, err := file.WriteAt(buf, headerSize-4)

	return err
}
