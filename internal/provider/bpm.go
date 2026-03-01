package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type BPMKey struct {
	BPM string
	Key string
}

var (
	bpmRegex  = regexp.MustCompile(`tempo of <span[^>]*>(\d+) BPM`)
	keyRegex  = regexp.MustCompile(`with a <span[^>]*>([A-G](?:♯|#|♭|b)?(?:/[A-G](?:♯|#|♭|b)?)?)</span> key`)
	modeRegex = regexp.MustCompile(`a  <span[^>]*>([a-z]+)</span> mode`)
)

func FetchBPM(ctx context.Context, httpClient *http.Client, artist, title, duration string) (BPMKey, error) {
	songURL, err := findSongURL(ctx, httpClient, artist, title, duration)
	if err != nil {
		return BPMKey{}, err
	}

	html, err := fetchBPMPage(ctx, httpClient, songURL)
	if err != nil {
		return BPMKey{}, err
	}

	return parseBPM(html)
}

func findSongURL(ctx context.Context, httpClient *http.Client, artist, title, duration string) (string, error) {
	const rootURL = "https://songbpm.com"

	values := neturl.Values{}
	values.Add("query", fmt.Sprintf("%s %s", artist, title))

	req, err := http.NewRequestWithContext(ctx, "POST", rootURL+"/searches", bytes.NewBufferString(values.Encode()))
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
		if foundDuration <= wantDuration-toleranceSec || foundDuration >= wantDuration+toleranceSec {
			return true
		}

		matchURL = sel.AttrOr("href", "")
		return false
	})

	if matchURL == "" {
		return "", fmt.Errorf("no data found")
	}

	return rootURL + matchURL, nil
}

func fetchBPMPage(ctx context.Context, httpClient *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

func parseBPM(html string) (BPMKey, error) {
	bpmMatch := bpmRegex.FindStringSubmatch(html)
	keyMatch := keyRegex.FindStringSubmatch(html)
	modeMatch := modeRegex.FindStringSubmatch(html)

	if len(bpmMatch) != 2 || len(keyMatch) != 2 || len(modeMatch) != 2 {
		return BPMKey{}, fmt.Errorf("no data found")
	}

	bpm := bpmMatch[1]
	key := strings.SplitN(keyMatch[1], "/", 2)[0]
	key = strings.ReplaceAll(key, "\u266f", "#")
	key = strings.ReplaceAll(key, "\u266d", "b")

	if modeMatch[1] == "minor" {
		key += "m"
	}

	return BPMKey{BPM: bpm, Key: key}, nil
}
