package deezer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// maxFallbackDepth caps the embedded FALLBACK chain. Deezer nests at most one
// level in practice, so this is slack rather than a budget.
const maxFallbackDepth = 3

// durationToleranceSec is how far two Deezer records of the same recording
// are allowed to disagree. DURATION is an integer rounding produced
// separately by each catalogue ingest, so one master can read 163 on one
// entry and 164 on another.
const durationToleranceSec = 1

// resolveFallback returns the media of a verified stand-in for track, or nil
// when none can be proved equivalent.
//
// Two places are tried in order: the FALLBACK object Deezer embeds in the
// response, then the track's ISRC on the public API, which finds duplicates
// Deezer did not link.
//
// The caller keeps the original entry's metadata, so a recovered track can
// carry the GAIN of a different release of the same recording. Taking the
// stand-in's metadata instead would mean the wrong cover and track number.
func (c *Client) resolveFallback(ctx context.Context, track *Track, quality string) *Media {
	// Without an ISRC there is no way to prove a candidate is the same
	// recording, so refuse before spending a single request.
	if normalizeISRC(track.ISRC) == "" {
		return nil
	}

	// seen keeps the ISRC lookup from re-offering a candidate the embedded
	// walk has already turned down.
	seen := map[string]bool{track.ID: true}

	for _, candidate := range embeddedCandidates(track, maxFallbackDepth, seen) {
		if media := c.mediaFrom(ctx, track, candidate, quality); media != nil {
			return media
		}
	}

	id := c.lookupByISRC(ctx, track.ISRC)
	if id == "" || seen[id] {
		return nil
	}

	// The page fetched here may carry a FALLBACK of its own; that chain is
	// not followed. One hop off the public API is enough, and more is
	// unbounded catalogue walking.
	resource, err := c.FetchResource(ctx, KindTrack, id)
	if err != nil {
		return nil
	}

	tracks := resource.Tracks()
	if len(tracks) == 0 {
		return nil
	}

	return c.mediaFrom(ctx, track, tracks[0], quality)
}

// mediaFrom returns the media of candidate when it is provably a stand-in for
// original, or nil when it is not or cannot be played.
//
// Candidates are verified against original, never against their parent in the
// fallback chain, which would let identity drift one hop at a time. This is
// also the only place a candidate becomes a Media, so the id paired with the
// audio is always the id whose token fetched it.
func (c *Client) mediaFrom(ctx context.Context, original, candidate *Track, quality string) *Media {
	if !sameRecording(original, candidate) {
		return nil
	}

	res, err := c.fetchMediaForToken(ctx, candidate.TrackToken, quality)
	if err != nil {
		return nil
	}

	return newMedia(candidate.ID, res)
}

// embeddedCandidates walks the FALLBACK chain and returns the candidates in
// order, stopping at maxDepth and at the first id already in seen. It adds
// every id it returns to seen, which is the only loop guard: a chain that
// loops back on itself ends there.
func embeddedCandidates(track *Track, maxDepth int, seen map[string]bool) []*Track {
	var candidates []*Track

	for next := track.Fallback; next != nil && len(candidates) < maxDepth; next = next.Fallback {
		if seen[next.ID] {
			break
		}
		seen[next.ID] = true
		candidates = append(candidates, next)
	}

	return candidates
}

// lookupByISRC asks the public API for the track carrying isrc and returns
// its id, or an empty string when there is no usable answer.
func (c *Client) lookupByISRC(ctx context.Context, isrc string) string {
	endpoint := "https://api.deezer.com/track/isrc:" + url.PathEscape(normalizeISRC(isrc))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ""
	}

	resp, err := c.Session.HTTPClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return parseISRCLookup(body)
}

// parseISRCLookup extracts a usable track id from a public API response, or
// an empty string when there is none.
//
// The public API reports both a miss and a rate limit as HTTP 200 with an
// error object in the body, so the status code alone says nothing.
func parseISRCLookup(body []byte) string {
	var res struct {
		// The public API sends the id as a JSON number where gw-light sends
		// a string. Decoding it into a string yields an empty value with no
		// error, which quietly disables the whole lookup.
		ID       json.Number `json:"id"`
		Readable bool        `json:"readable"`
		Error    *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return ""
	}

	// Readability is country and account scoped, so this is a cheap filter
	// and not the authority: the get_url call remains the real test.
	if res.Error != nil || !res.Readable {
		return ""
	}

	id := res.ID.String()
	if id == "0" {
		return ""
	}

	return id
}

// sameRecording reports whether candidate is provably the same recording as
// original.
//
// The ISRC is the identity. Duration is a second opinion, there to catch the
// case where a catalogue error puts one ISRC on two different masters.
func sameRecording(original, candidate *Track) bool {
	if candidate == nil || candidate.ID == "" || candidate.ID == original.ID || candidate.TrackToken == "" {
		return false
	}

	isrc := normalizeISRC(original.ISRC)
	if isrc == "" || isrc != normalizeISRC(candidate.ISRC) {
		return false
	}

	originalDuration, err := strconv.Atoi(original.Duration)
	if err != nil || originalDuration <= 0 {
		return false
	}

	candidateDuration, err := strconv.Atoi(candidate.Duration)
	if err != nil || candidateDuration <= 0 {
		return false
	}

	return max(originalDuration-candidateDuration, candidateDuration-originalDuration) <= durationToleranceSec
}

// isrcSeparators strips the punctuation catalogue entries sprinkle through an
// ISRC. The same code turns up bare, hyphenated and sometimes spaced, and two
// of them cannot be compared until they are punctuated the same way.
var isrcSeparators = strings.NewReplacer("-", "", " ", "")

func normalizeISRC(s string) string {
	return isrcSeparators.Replace(strings.ToUpper(s))
}
