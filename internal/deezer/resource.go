package deezer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Resource interface {
	GetType() string
	UnmarshalData(data []byte) error
	GetSongs() []*Song
	GetOutputPath(outputDir string) string
	GetTitle() string
}

func (s *Session) GetData(r Resource, id string) error {
	payload := map[string]interface{}{
		"nb":          10000,
		"start":       0,
		"playlist_id": id,
		"alb_id":      id,
		"lang":        "en",
		"tab":         0,
		"tags":        true,
		"header":      true,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://www.deezer.com/ajax/gw-light.php?method=deezer.page%s&input=3&api_version=1.0&api_token=%s", r.GetType(), s.APIToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	if strings.Contains(string(body), `"DATA_ERROR":"playlist::getData"`) {
		return fmt.Errorf("invalid playlist ID")
	}
	if strings.Contains(string(body), `"DATA_ERROR":"album::getData"`) {
		return fmt.Errorf("invalid album ID")
	}
	if strings.Contains(string(body), `"results":{}`) {
		return fmt.Errorf("unexpected response")
	}

	return r.UnmarshalData(body)
}
