package deezer

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mathismqn/godeez/internal/crypto"
)

type Media struct {
	Errors []MediaError `json:"errors"`
	Data   []struct {
		Media []struct {
			Type    string   `json:"media_type"`
			Cipher  Cipher   `json:"cipher"`
			Format  string   `json:"format"`
			Sources []Source `json:"sources"`
		}
	} `json:"data"`
}

type MediaError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Cipher struct {
	Type string `json:"type"`
}

type Source struct {
	URL      string `json:"url"`
	Provider string `json:"provider"`
}

const ChunkSize = 2048

func (m *Media) Download(filename, songID string) error {
	url := m.Data[0].Media[0].Sources[0].URL

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	key := crypto.GetBlowfishKey(songID)
	buffer := make([]byte, ChunkSize)

	for chunk := 0; ; chunk++ {
		totalRead := 0
		for totalRead < ChunkSize {
			n, err := resp.Body.Read(buffer[totalRead:])
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				return err
			}

			if n > 0 {
				totalRead += n
			}
		}

		if totalRead == 0 {
			break
		}

		if chunk%3 == 0 && totalRead == ChunkSize {
			buffer, err = crypto.DecryptBlowfish(buffer, key)
			if err != nil {
				return err
			}
		}

		_, err = file.Write(buffer[:totalRead])
		if err != nil {
			return err
		}

		if totalRead < ChunkSize {
			break
		}
	}

	return nil
}
