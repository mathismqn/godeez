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
	ID           string `json:"SNG_ID"`
	Artist       string `json:"ART_NAME"`
	Title        string `json:"SNG_TITLE"`
	Version      string `json:"VERSION"`
	Cover        string `json:"ALB_PICTURE"`
	Contributors struct {
		MainArtists []string `json:"main_artist"`
		Composers   []string `json:"composer"`
		Authors     []string `json:"author"`
	} `json:"SNG_CONTRIBUTORS"`
	Duration    string `json:"DURATION"`
	Gain        string `json:"GAIN"`
	ISRC        string `json:"ISRC"`
	TrackNumber string `json:"TRACK_NUMBER"`
	TrackToken  string `json:"TRACK_TOKEN"`
}

func (s *Song) GetMediaData(quality string) (*Media, error) {
	var formats string

	switch quality {
	case "mp3_128":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	case "mp3_320":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"MP3_320"}]`
	case "flac":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"FLAC"}]`
	case "best":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"FLAC"},{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	}

	reqBody := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":%s}],"track_tokens":["%s"]}`, config.Cfg.LicenseToken, formats, s.TrackToken)
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
	if len(media.Data) > 0 && len(media.Data[0].Errors) > 0 {
		if media.Data[0].Errors[0].Code == 2002 {
			return nil, fmt.Errorf("invalid track token")
		}

		return nil, fmt.Errorf("%s", media.Data[0].Errors[0].Message)
	}

	return &media, nil
}

func (s *Song) GetCoverImage() ([]byte, error) {
	url := fmt.Sprintf("https://e-cdn-images.dzcdn.net/images/cover/%s/500x500-000000-80-0-0.jpg", s.Cover)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
