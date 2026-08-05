package download

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/mathismqn/godeez/internal/deezer"
)

const (
	chunkSize   = 2048
	partPattern = ".godeez-*.part"
)

func sweepPartFiles(dir string) {
	matches, err := filepath.Glob(filepath.Join(dir, partPattern))
	if err != nil {
		return
	}
	for _, match := range matches {
		os.Remove(match)
	}
}

func (d *Downloader) streamToFile(ctx context.Context, stream io.ReadCloser, outputPath string, key []byte) error {
	defer stream.Close()

	file, err := os.CreateTemp(filepath.Dir(outputPath), partPattern)
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

	buffer := make([]byte, chunkSize)
	for chunk := 0; ; chunk++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		totalRead := 0
		for totalRead < chunkSize {
			n, err := stream.Read(buffer[totalRead:])
			totalRead += n
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
		}

		if totalRead == 0 {
			break
		}

		if chunk%3 == 0 && totalRead == chunkSize {
			buffer, err = deezer.DecryptBlowfish(buffer, key)
			if err != nil {
				return err
			}
		}

		if _, err = file.Write(buffer[:totalRead]); err != nil {
			return err
		}

		if totalRead < chunkSize {
			break
		}
	}

	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, outputPath); err != nil {
		return err
	}
	done = true

	return nil
}
