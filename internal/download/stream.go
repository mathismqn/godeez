package download

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/mathismqn/godeez/internal/deezer"
)

const chunkSize = 2048

func (d *Downloader) streamToFile(ctx context.Context, stream io.ReadCloser, outputPath string, key []byte) error {
	defer stream.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

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

	return nil
}
