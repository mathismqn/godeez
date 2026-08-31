package deezer

import (
	"context"
	"slices"
	"strings"
)

// maxFallbackDepth caps the embedded FALLBACK chain. Deezer nests at most one
// level in practice, so this is slack rather than a budget.
const maxFallbackDepth = 3

// resolveFallback returns the media of a verified stand-in for track, or nil
// when none can be verified.
//
// Candidates come from the FALLBACK object Deezer embeds in the response and
// from nowhere else. Deezer sets that field when an entry plays from another
// id and the official client simply follows it, so a track with no usable
// chain is one the app cannot play either. Looking the catalogue up by the
// original ISRC adds nothing: it finds only duplicates Deezer has already
// linked, because its own linking survives the re-registration that gives a
// re-release a new code.
//
// The caller keeps the original entry's metadata, so a recovered track can
// carry the GAIN of a different release of the same recording. Taking the
// stand-in's metadata instead would mean the wrong track number. The cover is
// the one thing borrowed from it, and only when the original has none; see
// coverCandidates.
func (c *Client) resolveFallback(ctx context.Context, track *Track, quality string) *Media {
	for _, candidate := range embeddedCandidates(track, maxFallbackDepth) {
		if media := c.mediaFrom(ctx, track, candidate, quality); media != nil {
			return media
		}
	}

	return nil
}

// mediaFrom returns the media of candidate when it verifies as a stand-in for
// original, or nil when it does not or cannot be played.
//
// Candidates are verified against original, never against their parent in the
// fallback chain, which would let identity drift one hop at a time. This is
// also the only place a candidate becomes a Media, so the id paired with the
// audio is always the id whose token fetched it.
func (c *Client) mediaFrom(ctx context.Context, original, candidate *Track, quality string) *Media {
	if !linkedStandIn(original, candidate) {
		return nil
	}

	res, err := c.fetchMediaForToken(ctx, candidate.TrackToken, quality)
	if err != nil {
		return nil
	}

	return newMedia(string(candidate.ID), res)
}

// coverCandidates returns the cover ids worth trying for track, best first:
// its own, then those of the stand-ins Deezer links it to.
//
// A dead entry carries no cover of its own, either as an empty ALB_PICTURE or
// as missingCoverMD5 spelled out in the field, while the stand-in that
// replaces it still carries a sleeve. Candidates are held to the same
// linkedStandIn check as the audio, so the artwork comes from another release
// of the same recording rather than from a different song.
func coverCandidates(track *Track) []string {
	var covers []string
	add := func(cover string) {
		if usableCover(cover) && !slices.Contains(covers, cover) {
			covers = append(covers, cover)
		}
	}

	add(track.Cover)
	for _, candidate := range embeddedCandidates(track, maxFallbackDepth) {
		if usableCover(candidate.Cover) && linkedStandIn(track, candidate) {
			add(candidate.Cover)
		}
	}

	return covers
}

// embeddedCandidates walks the FALLBACK chain and returns the candidates in
// order, stopping at maxDepth and at the first id it has already returned.
// The ids it has returned are the only loop guard: a chain that loops back on
// itself, or back to the original, ends there.
func embeddedCandidates(track *Track, maxDepth int) []*Track {
	seen := map[Number]bool{track.ID: true}

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

// linkedStandIn reports whether candidate is an acceptable stand-in for
// original.
//
// Each identifier can vouch for a candidate but neither can veto one, since a
// mismatch is not evidence of a different recording: a re-release is
// registered under a new ISRC while staying the same master, and Deezer keeps
// pointing FALLBACK at it. PRODUCT_TRACK_ID is asked first because it names
// the recording rather than this listing of it, so it carries across every
// release of one master. The ISRC adds no coverage in practice but is a public
// standard rather than a field of Deezer's own gateway, so it still answers if
// PRODUCT_TRACK_ID stops being sent.
func linkedStandIn(original, candidate *Track) bool {
	if candidate == nil || candidate.ID == "" || candidate.ID == original.ID || candidate.TrackToken == "" {
		return false
	}

	if original.ProductID != "" && original.ProductID == candidate.ProductID {
		return true
	}

	return original.ISRC != "" && strings.EqualFold(original.ISRC, candidate.ISRC)
}
