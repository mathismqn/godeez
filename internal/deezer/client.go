package deezer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	Session *Session
}

func NewClient(ctx context.Context, arlCookie string) (*Client, error) {
	session, err := resolveSession(ctx, arlCookie)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return &Client{
		Session: session,
	}, nil
}

func resolveSession(ctx context.Context, arlCookie string) (*Session, error) {
	if arlCookie != "" {
		return authenticate(ctx, arlCookie)
	}

	var session *Session
	validate := func(ctx context.Context, arl string) error {
		s, err := authenticate(ctx, arl)
		if err != nil {
			return err
		}
		session = s

		return nil
	}

	arl, err := resolveARL(ctx, validate)
	if err != nil {
		return nil, err
	}

	if session == nil {
		session, err = authenticate(ctx, arl)
		if err != nil {
			return nil, err
		}
	}

	return session, nil
}

func (c *Client) FetchResource(ctx context.Context, kind Kind, id string) (Resource, error) {
	resource, err := kind.newResource()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"nb":     10000,
		"start":  0,
		"lang":   "en",
		"tab":    0,
		"tags":   true,
		"header": true,
	}
	payload[kind.idKey()] = id

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://www.deezer.com/ajax/gw-light.php?method=deezer.page%s&input=3&api_version=1.0&api_token=%s", kind.pageMethod(), c.Session.APIToken)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyStr := string(body)
	for _, check := range []struct {
		marker string
		errMsg string
	}{
		{`"DATA_ERROR":"playlist::getData"`, "invalid playlist ID"},
		{`"DATA_ERROR":"album::getData"`, "invalid album ID"},
		{`"DATA_ERROR":"artist::getData"`, "invalid artist ID"},
		{`"DATA_ERROR":"song::getData"`, "invalid track ID"},
	} {
		if strings.Contains(bodyStr, check.marker) {
			return nil, fmt.Errorf("%s", check.errMsg)
		}
	}

	if strings.Contains(bodyStr, `"results":{}`) {
		return nil, fmt.Errorf("unexpected response")
	}

	if err := resource.Unmarshal(body); err != nil {
		return nil, err
	}

	return resource, nil
}

func (c *Client) FetchMedia(ctx context.Context, track *Track, quality string) (*Media, error) {
	qualityFormats := map[string]string{
		"mp3_128": `[{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`,
		"mp3_320": `[{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`,
		"flac":    `[{"cipher":"BF_CBC_STRIPE","format":"FLAC"},{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`,
	}

	reqBody := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":%s}],"track_tokens":["%s"]}`, c.Session.LicenseToken, qualityFormats[quality], track.TrackToken)
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
	if err := json.Unmarshal(body, &media); err != nil {
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

	if len(media.Data) == 0 || len(media.Data[0].Media) == 0 || len(media.Data[0].Media[0].Sources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	return &media, nil
}

func (c *Client) FetchCoverImage(ctx context.Context, track *Track) ([]byte, error) {
	url := fmt.Sprintf("https://e-cdn-images.dzcdn.net/images/cover/%s/500x500-000000-80-0-0.jpg", track.Cover)
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

func (c *Client) GetMediaStream(ctx context.Context, media *Media) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", media.GetURL(), nil)
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
