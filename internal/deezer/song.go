package deezer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mathismqn/godeez/internal/config"
)

type Song struct {
	ID         string `json:"SNG_ID"`
	ArtistName string `json:"ART_NAME"`
	Title      string `json:"SNG_TITLE"`
	Version    string `json:"VERSION"`
	TrackToken string `json:"TRACK_TOKEN"`
}

func (s *Song) GetMediaData() (*Media, error) {
	reqBody := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":[{"cipher":"BF_CBC_STRIPE","format":"FLAC"}]}],"track_tokens":["%s"]}`, config.Cfg.LicenseToken, s.TrackToken)

	resp, err := http.Post("https://media.deezer.com/v1/get_url", "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var media Media
	err = json.Unmarshal(body, &media)
	if err != nil {
		return nil, err
	}

	if len(media.Errors) > 0 {
		if media.Errors[0].Code == 1000 {
			return nil, fmt.Errorf("invalid license token")
		}

		return nil, fmt.Errorf("%s", media.Errors[0].Message)
	}

	return &media, nil
}
