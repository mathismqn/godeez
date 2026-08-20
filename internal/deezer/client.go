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
//
// Entries with no streaming rights of their own are resolved to a verified
// duplicate rather than failing; see fallback.go.
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

// errTrackUnavailable reports that the media server refused an entry because
// it carries no streaming rights. The track token is not malformed and the
// session is fine.
//
// Deezer says this two ways and fetchMediaForToken maps both onto this one
// sentinel: media error code 2002, and a 200 whose media array is empty. A
// third way belongs here too, since this is what FetchMedia tests before it
// looks for a stand-in.
var errTrackUnavailable = errors.New("track is not available for streaming")

// FetchMedia resolves track to playable media at the requested quality.
//
// An entry with no streaming rights of its own is played from a verified
// duplicate when Deezer publishes one, so the returned Media may carry a
// different track id than the one asked for. The substitution is silent
// because the recording is the same, checked against the original's artist,
// title and duration; only the bytes come from elsewhere.
func (c *Client) FetchMedia(ctx context.Context, track *Track, quality string) (*Media, error) {
	res, err := c.fetchMediaForToken(ctx, track.TrackToken, quality)
	if err == nil {
		return newMedia(track.ID, res), nil
	}

	if !errors.Is(err, errTrackUnavailable) {
		return nil, err
	}

	if media := c.resolveFallback(ctx, track, quality); media != nil {
		return media, nil
	}

	// resolveFallback reports every failure the same way, so a candidate
	// fetch cancelled mid-flight would otherwise surface as an unavailable
	// track and let the download loop keep going after an interrupt.
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}

	return nil, err
}

// fetchMediaForToken performs the get_url call for a single track token.
//
// Each quality maps to an ordered fallback chain, so asking for flac on a
// track that has none yields mp3_320 instead of an error; callers compare
// Media.Format against what they asked for to detect a downgrade. There is
// no wav entry because Deezer does not serve wav: the download package
// requests flac and converts locally.
//
// A 400 is accepted alongside 200 because the gateway uses it to return a
// structured error payload that is more useful than the status code.
//
// The fallback resolver retries this call with another token, so it must
// stay free of fallback logic of its own or the two would recurse.
func (c *Client) fetchMediaForToken(ctx context.Context, trackToken, quality string) (*mediaResponse, error) {
	var formats string
	switch quality {
	case "mp3_128":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	case "mp3_320":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	case "flac":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"FLAC"},{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	}

	reqBody := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":%s}],"track_tokens":["%s"]}`, c.Session.licenseToken, formats, trackToken)
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

	var res mediaResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if len(res.Errors) > 0 {
		if res.Errors[0].Code == 1000 {
			return nil, errors.New("invalid license token")
		}
		return nil, errors.New(res.Errors[0].Message)
	}

	if len(res.Data) > 0 && len(res.Data[0].Errors) > 0 {
		if res.Data[0].Errors[0].Code == 2002 {
			return nil, errTrackUnavailable
		}
		return nil, errors.New(res.Data[0].Errors[0].Message)
	}

	if len(res.Data) == 0 {
		return nil, errors.New("no sources found")
	}

	// The empty media array is the second way a track with no streaming
	// rights is refused; see errTrackUnavailable.
	if len(res.Data[0].Media) == 0 {
		return nil, errTrackUnavailable
	}

	if len(res.Data[0].Media[0].Sources) == 0 {
		return nil, errors.New("no sources found")
	}

	return &res, nil
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
// body and must close it, and decrypt it with media.Key.
//
// The session client is copied so its timeout can be cleared for this
// request: the session timeout is sized for short API calls and would abort
// a long track transfer. Cancellation is left to ctx.
func (c *Client) MediaStream(ctx context.Context, media *Media) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, media.url, nil)
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
