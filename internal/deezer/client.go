// Package deezer talks to Deezer's private endpoints: the gw-light web API,
// the Android mobile gateway used for email and password login, and the
// media servers that hand out encrypted audio streams. None of it is
// documented or supported by Deezer, so the request shapes, the error
// markers matched in response bodies and the crypto constants in this
// package were all derived from the official clients and can break without
// warning.
//
// A Client wraps an authenticated Session and fetches a Resource, which is
// one of Album, Playlist, Artist or Single. Audio is served Blowfish
// encrypted; see blowfish.go for the key derivation and the download
// package for the stripe pattern that undoes it.
package deezer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	Session *Session
}

// NewClient authenticates with Deezer and returns a client bound to the
// resulting session. An empty arlCookie falls back to the credentials held
// in the system keyring.
func NewClient(ctx context.Context, arlCookie string) (*Client, error) {
	session, err := resolveSession(ctx, arlCookie)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return &Client{
		Session: session,
	}, nil
}

// resolveSession authenticates with arlCookie when one is supplied, and
// otherwise falls back to the stored credentials.
//
// The validate callback handed to resolveARL is the real authentication, not
// a separate probe, so a stored ARL that still works is not sent twice.
// session is therefore only nil here when resolveARL had to log in again to
// mint a fresh ARL.
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

// FetchResource fetches the page for the given kind and id and decodes it
// into the matching Resource implementation.
//
// gw-light answers 200 even for an unknown id and reports the failure inside
// the JSON, so bad ids have to be detected by matching markers in the body
// rather than by reading the status code. The nb parameter is set far above
// any real tracklist length to pull an entire resource in one request and
// avoid paging.
func (c *Client) FetchResource(ctx context.Context, kind Kind, id string) (Resource, error) {
	resource, err := kind.newResource()
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
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

	url := fmt.Sprintf("https://www.deezer.com/ajax/gw-light.php?method=deezer.page%s&input=3&api_version=1.0&api_token=%s", kind.pageMethod(), c.Session.apiToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	resp, err := c.Session.HTTPClient.Do(req)
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
			return nil, errors.New(check.errMsg)
		}
	}

	if strings.Contains(bodyStr, `"results":{}`) {
		return nil, errors.New("unexpected response")
	}

	if err := resource.decode(body); err != nil {
		return nil, err
	}

	return resource, nil
}

// FetchMedia resolves a playable source URL for track at the requested
// quality.
//
// Each quality maps to an ordered fallback chain, so asking for flac on a
// track that has none yields mp3_320 instead of an error; callers compare
// Media.Format against what they asked for to detect a downgrade. There is
// no wav entry because Deezer does not serve wav: the download package
// requests flac and converts locally.
//
// A 400 is accepted alongside 200 because the gateway uses it to return a
// structured error payload that is more useful than the status code.
func (c *Client) FetchMedia(ctx context.Context, track *Track, quality string) (*Media, error) {
	qualityFormats := map[string]string{
		"mp3_128": `[{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`,
		"mp3_320": `[{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`,
		"flac":    `[{"cipher":"BF_CBC_STRIPE","format":"FLAC"},{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`,
	}

	reqBody := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":%s}],"track_tokens":["%s"]}`, c.Session.licenseToken, qualityFormats[quality], track.TrackToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://media.deezer.com/v1/get_url", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return nil, err
	}

	resp, err := c.Session.HTTPClient.Do(req)
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
			return nil, errors.New("invalid license token")
		}
		return nil, errors.New(media.Errors[0].Message)
	}

	if len(media.Data) > 0 && len(media.Data[0].Errors) > 0 {
		if media.Data[0].Errors[0].Code == 2002 {
			return nil, errors.New("invalid track token")
		}
		return nil, errors.New(media.Data[0].Errors[0].Message)
	}

	if len(media.Data) == 0 || len(media.Data[0].Media) == 0 || len(media.Data[0].Media[0].Sources) == 0 {
		return nil, errors.New("no sources found")
	}

	return &media, nil
}

func (c *Client) FetchCoverImage(ctx context.Context, track *Track) ([]byte, error) {
	url := fmt.Sprintf("https://e-cdn-images.dzcdn.net/images/cover/%s/500x500-000000-80-0-0.jpg", track.Cover)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Session.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// MediaStream opens the audio stream for media. The caller owns the returned
// body and must close it.
//
// The session client is copied so its timeout can be cleared for this
// request: the session timeout is sized for short API calls and would abort
// a long track transfer. Cancellation is left to ctx.
func (c *Client) MediaStream(ctx context.Context, media *Media) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, media.URL(), nil)
	if err != nil {
		return nil, err
	}

	streamingClient := *c.Session.HTTPClient
	streamingClient.Timeout = 0

	resp, err := streamingClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
