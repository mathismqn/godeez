package deezer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// missingCoverMD5 is the MD5 of the empty string. Deezer's image CDN redirects
// every empty or unknown cover id there and serves a grey placeholder.
const missingCoverMD5 = "d41d8cd98f00b204e9800998ecf8427e"

// usableCover reports whether cover could name real artwork. An entry with
// none leaves ALB_PICTURE empty or spells missingCoverMD5 out in it, and both
// only ever resolve to the placeholder, so neither is worth a request.
func usableCover(cover string) bool {
	return cover != "" && cover != missingCoverMD5
}

// ErrPlaceholderCover reports that Deezer has no artwork for the track. It
// comes back alongside the placeholder image rather than instead of it, so a
// caller that would rather embed something than nothing still can.
var ErrPlaceholderCover = errors.New("no cover art available")

// FetchCoverImage returns the artwork for track, trying the covers
// coverCandidates offers until one of them is real artwork.
//
// A cover that does not exist cannot be told apart by status code, since the
// CDN redirects to the placeholder and serves it with a 200. A track with no
// artwork anywhere still yields an image, by asking for the placeholder
// outright when no candidate has already produced it.
func (c *Client) FetchCoverImage(ctx context.Context, track *Track) ([]byte, error) {
	var placeholder []byte

	for _, cover := range coverCandidates(track) {
		image, err := c.fetchCover(ctx, cover)
		if err == nil {
			return image, nil
		}
		if errors.Is(err, ErrPlaceholderCover) {
			placeholder = image
		}
	}

	if placeholder == nil {
		return c.fetchCover(ctx, missingCoverMD5)
	}

	return placeholder, ErrPlaceholderCover
}

// fetchCover downloads one cover id, returning the placeholder with
// ErrPlaceholderCover when that is where the CDN sent the request.
func (c *Client) fetchCover(ctx context.Context, cover string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, coverURL(cover), nil)
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

	image, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if isPlaceholderCoverURL(resp.Request.URL) {
		return image, ErrPlaceholderCover
	}

	return image, nil
}

func coverURL(cover string) string {
	return fmt.Sprintf("https://e-cdn-images.dzcdn.net/images/cover/%s/500x500-000000-80-0-0.jpg", cover)
}

// isPlaceholderCoverURL reports whether u is where the CDN sends a request for
// a cover that does not exist. Pass the URL of the response, which net/http
// points at the last request made and so reflects the redirect.
func isPlaceholderCoverURL(u *url.URL) bool {
	return u != nil && strings.Contains(u.Path, missingCoverMD5)
}
