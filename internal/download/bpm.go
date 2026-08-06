package download

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// songbpm.com publishes the tempo and key as prose rather than as structured
// data, so these match the surrounding sentence instead of a CSS selector.
// The attribute wildcards absorb the utility classes the site regenerates on
// every deploy, but the wording itself is load bearing: if the sentence
// changes, the lookup starts returning no data. The double space in modeRegex
// is present in the real markup and is not a typo.
//
// The key pattern accepts both the typographic accidentals the page renders
// and their ASCII equivalents, since which one appears varies by track.
var (
	bpmRegex  = regexp.MustCompile(`tempo of <span[^>]*>(\d+) BPM`)
	keyRegex  = regexp.MustCompile(`with a <span[^>]*>([A-G](?:♯|#|♭|b)?(?:/[A-G](?:♯|#|♭|b)?)?)</span> key`)
	modeRegex = regexp.MustCompile(`a  <span[^>]*>([a-z]+)</span> mode`)
)

func fetchBPM(ctx context.Context, httpClient *http.Client, artist, title, duration string) (bpmKey, error) {
	trackURL, err := findTrackURL(ctx, httpClient, artist, title, duration)
	if err != nil {
		return bpmKey{}, err
	}

	html, err := fetchBPMPage(ctx, httpClient, trackURL)
	if err != nil {
		return bpmKey{}, err
	}

	return parseBPM(html)
}

// findTrackURL searches songbpm.com and returns the page for the track.
//
// Artist and title alone are not enough to identify a track, since the search
// happily returns remixes, live versions and covers under the same names.
// Duration is used as the tiebreaker, with a couple of seconds of tolerance
// to absorb the disagreement between Deezer's rounding and songbpm's. No
// match within tolerance is treated as not found rather than guessed at,
// because a wrong BPM is worse than a missing one.
func findTrackURL(ctx context.Context, httpClient *http.Client, artist, title, duration string) (string, error) {
	const rootURL = "https://songbpm.com"

	values := url.Values{}
	values.Add("query", fmt.Sprintf("%s %s", artist, title))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rootURL+"/searches", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", rootURL)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	wantDuration, err := strconv.Atoi(duration)
	if err != nil {
		return "", fmt.Errorf("invalid duration: %w", err)
	}

	lowerTitle := strings.ToLower(title)
	lowerArtist := strings.ToLower(artist)

	var matchURL string
	doc.Find("a.flex.flex-col").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		text := strings.ToLower(sel.Text())
		if !strings.Contains(text, lowerTitle) || !strings.Contains(text, lowerArtist) {
			return true
		}

		durationStr := strings.TrimSpace(sel.Find("div.flex-1.flex-col.items-center").Eq(1).Find("span.text-2xl").Text())
		parts := strings.Split(durationStr, ":")
		if len(parts) != 2 {
			return true
		}
		minutes, err := strconv.Atoi(parts[0])
		if err != nil {
			return true
		}
		seconds, err := strconv.Atoi(parts[1])
		if err != nil {
			return true
		}

		const toleranceSec = 2
		foundDuration := minutes*60 + seconds
		if foundDuration < wantDuration-toleranceSec || foundDuration > wantDuration+toleranceSec {
			return true
		}

		matchURL = sel.AttrOr("href", "")
		return false
	})

	if matchURL == "" {
		return "", errors.New("no data found")
	}

	return rootURL + matchURL, nil
}

func fetchBPMPage(ctx context.Context, httpClient *http.Client, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// parseBPM extracts the tempo and musical key from a track page.
//
// All three patterns must match: a page with a tempo but no key is treated as
// no data, since a half filled tag is not worth writing. Enharmonic keys are
// published as pairs like "C#/Db" and only the first spelling is kept, the
// accidentals are folded to ASCII for tag compatibility, and a minor mode is
// encoded with a trailing "m" to match the convention DJ software expects.
func parseBPM(html string) (bpmKey, error) {
	bpmMatch := bpmRegex.FindStringSubmatch(html)
	keyMatch := keyRegex.FindStringSubmatch(html)
	modeMatch := modeRegex.FindStringSubmatch(html)

	if len(bpmMatch) != 2 || len(keyMatch) != 2 || len(modeMatch) != 2 {
		return bpmKey{}, errors.New("no data found")
	}

	bpm := bpmMatch[1]
	key := strings.SplitN(keyMatch[1], "/", 2)[0]
	key = strings.ReplaceAll(key, "\u266f", "#")
	key = strings.ReplaceAll(key, "\u266d", "b")

	if modeMatch[1] == "minor" {
		key += "m"
	}

	return bpmKey{BPM: bpm, Key: key}, nil
}
