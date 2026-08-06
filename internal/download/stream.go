package download

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/mathismqn/godeez/internal/fsutil"
)

// chunkSize is the stripe width Deezer encrypts with. It is fixed by the
// BF_CBC_STRIPE cipher named in the media request and is not tunable: reading
// in any other unit would misalign the stripe pattern and corrupt the output.
const chunkSize = 2048

// sweepPartFiles deletes leftover .part files in dir. Failures are ignored
// because this is opportunistic cleanup, and refusing to download because a
// stale temp file could not be removed would be worse than leaving it.
func sweepPartFiles(dir string) {
	matches, err := filepath.Glob(filepath.Join(dir, fsutil.PartPattern))
	if err != nil {
		return
	}
	for _, match := range matches {
		os.Remove(match)
	}
}

// streamToFile writes the decrypted stream to outputPath.
//
// The download lands in a temporary file first and is only renamed into place
// once it is complete, so an interrupted run never leaves a truncated file
// sitting at the real path where it would look like a finished download.
func (d *Downloader) streamToFile(ctx context.Context, stream io.ReadCloser, outputPath string, key []byte) error {
	tmpPath, err := d.streamToTempFile(ctx, stream, filepath.Dir(outputPath), key)
	if err != nil {
		return err
	}

	if err := os.Rename(tmpPath, outputPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// streamToTempFile decrypts the stream into a .part file in dir and returns
// its path. The caller owns the file from that point on. It closes stream.
//
// The temp file is created in the destination directory rather than the
// system temp dir so the caller's rename stays on one filesystem and is
// therefore atomic.
func (d *Downloader) streamToTempFile(ctx context.Context, stream io.ReadCloser, dir string, key []byte) (string, error) {
	defer stream.Close()

	file, err := os.CreateTemp(dir, fsutil.PartPattern)
	if err != nil {
		return "", err
	}
	tmpPath := file.Name()
	// done stays false until the file is fully written and closed, so every
	// early return below removes the partial file instead of orphaning it.
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
			return "", ctx.Err()
		default:
		}

		// Read until the chunk is full rather than trusting a single Read.
		// A short read is legal and common on a network stream, and treating
		// one as a chunk boundary would shift every following chunk out of
		// step with the stripe pattern.
		totalRead := 0
		for totalRead < chunkSize {
			n, err := stream.Read(buffer[totalRead:])
			totalRead += n
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return "", err
			}
		}

		if totalRead == 0 {
			break
		}

		// Deezer encrypts only every third chunk and leaves the other two in
		// the clear, which is what BF_CBC_STRIPE means. A trailing partial
		// chunk is never encrypted even when its index is a multiple of three,
		// hence the length check: decrypting it would corrupt the end of the
		// file.
		if chunk%3 == 0 && totalRead == chunkSize {
			buffer, err = deezer.DecryptBlowfish(buffer, key)
			if err != nil {
				return "", err
			}
		}

		if _, err = file.Write(buffer[:totalRead]); err != nil {
			return "", err
		}

		if totalRead < chunkSize {
			break
		}
	}

	if err := file.Sync(); err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	done = true

	return tmpPath, nil
}
