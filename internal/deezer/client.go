package deezer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mathismqn/godeez/internal/config"
)

type Client struct {
	AppConfig *config.Config
	Session   *Session
}

func NewClient(ctx context.Context, appConfig *config.Config) (*Client, error) {
	session, err := Authenticate(ctx, appConfig.ArlCookie)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return &Client{
		AppConfig: appConfig,
		Session:   session,
	}, nil
}

func (c *Client) FetchResource(ctx context.Context, resource Resource, id string) error {
	payload := map[string]interface{}{
		"nb":     10000,
		"start":  0,
		"lang":   "en",
		"tab":    0,
		"tags":   true,
		"header": true,
	}
	switch r := resource.(type) {
	case *Playlist:
		payload["playlist_id"] = id
	case *Album:
		payload["alb_id"] = id
	case *Artist:
		payload["art_id"] = id
	default:
		return fmt.Errorf("unsupported resource type: %T", r)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://www.deezer.com/ajax/gw-light.php?method=deezer.page%s&input=3&api_version=1.0&api_token=%s", resource.GetType(), c.Session.APIToken)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	resp, err := c.Session.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Contains(string(body), `"DATA_ERROR":"playlist::getData"`) {
		return fmt.Errorf("invalid playlist ID")
	}
	if strings.Contains(string(body), `"DATA_ERROR":"album::getData"`) {
		return fmt.Errorf("invalid album ID")
	}
	if strings.Contains(string(body), `"DATA_ERROR":"artist::getData"`) {
		return fmt.Errorf("invalid artist ID")
	}
	if strings.Contains(string(body), `"results":{}`) {
		return fmt.Errorf("unexpected response")
	}

	return resource.Unmarshal(body)
}

func (c *Client) FetchMedia(ctx context.Context, song *Song, quality string) (*Media, error) {
	var formats string

	switch quality {
	case "mp3_128":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	case "mp3_320":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	case "flac":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"FLAC"},{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	}

	reqBody := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":%s}],"track_tokens":["%s"]}`, c.Session.LicenseToken, formats, song.TrackToken)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://media.deezer.com/v1/get_url", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return nil, err
	}

	resp, err := c.Session.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

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

func (c *Client) FetchCoverImage(ctx context.Context, song *Song) ([]byte, error) {
	url := fmt.Sprintf("https://e-cdn-images.dzcdn.net/images/cover/%s/500x500-000000-80-0-0.jpg", song.Cover)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Session.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) GetMediaStream(ctx context.Context, media *Media, songID string) (io.ReadCloser, error) {
	url, err := media.GetURL()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	streamingClient := *c.Session.HttpClient
	streamingClient.Timeout = 0

	resp, err := streamingClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
