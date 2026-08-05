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
	headerSize       = 44
	formatPCM        = 1
	ctxCheckInterval = 64
	maxDataSize      = math.MaxUint32 - (headerSize - 8)
)

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
