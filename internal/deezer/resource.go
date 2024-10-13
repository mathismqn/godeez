package deezer

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type Resource interface {
	GetURL(id string) string
	UnmarshalData(data []byte) error
	GetSongs() []*Song
	GetOutputPath(outputDir string) string
	GetTitle() string
}

func GetData(r Resource, id string) error {
	url := r.GetURL(id)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("resource not found")
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`window\.__DZR_APP_STATE__ = (\{.*\})`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		return fmt.Errorf("error parsing response")
	}

	return r.UnmarshalData([]byte(matches[1]))
}
