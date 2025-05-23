package bpm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Metrics struct {
	BPM string
	Key string
}

func FetchMetrics(ctx context.Context, httpClient *http.Client, artist, title, duration string) (*Metrics, error) {
	url, err := findSongURL(ctx, httpClient, artist, title, duration)
	if err != nil {
		return nil, err
	}

	html, err := fetchPage(ctx, httpClient, url)
	if err != nil {
		return nil, err
	}

	return parseMetrics(html)
}

func findSongURL(ctx context.Context, httpClient *http.Client, artist, title, duration string) (string, error) {
	rootUrl := "https://songbpm.com"
	reqUrl := rootUrl + "/searches"

	values := url.Values{}
	values.Add("query", fmt.Sprintf("%s %s", artist, title))

	req, err := http.NewRequestWithContext(ctx, "POST", reqUrl, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://songbpm.com")

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

	var (
		found bool
		url   string
	)

	doc.Find("a.flex.flex-col").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		lowerSelection := strings.ToLower(selection.Text())
		lowerTitle := strings.ToLower(title)
		lowerArtist := strings.ToLower(artist)
		if !strings.Contains(lowerSelection, lowerTitle) || !strings.Contains(lowerSelection, lowerArtist) {
			return true
		}

		durationStr := strings.TrimSpace(selection.Find("div.flex-1.flex-col.items-center").Eq(1).Find("span.text-2xl").Text())
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

		foundDuration := minutes*60 + seconds
		duration, err := strconv.Atoi(duration)
		if err != nil {
			return true
		}

		if foundDuration <= (duration-2) || foundDuration >= (duration+2) {
			return true
		}

		url = selection.AttrOr("href", "")
		found = true

		return false
	})

	if !found {
		return "", fmt.Errorf("no data found")
	}

	return rootUrl + url, nil
}

func fetchPage(ctx context.Context, httpClient *http.Client, url string) (string, error) {
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

func parseMetrics(html string) (*Metrics, error) {
	bpmRegex := regexp.MustCompile(`tempo of <span[^>]*>(\d+) BPM`)
	bpmMatch := bpmRegex.FindStringSubmatch(html)

	keyRegex := regexp.MustCompile(`with a <span[^>]*>([A-G](?:♯|#|♭|b)?(?:/[A-G](?:♯|#|♭|b)?)?)</span> key`)
	keyMatch := keyRegex.FindStringSubmatch(html)

	modeRegex := regexp.MustCompile(`a  <span[^>]*>([a-z]+)</span> mode`)
	modeMatch := modeRegex.FindStringSubmatch(html)

	if len(bpmMatch) != 2 || len(keyMatch) != 2 || len(modeMatch) != 2 {
		return nil, fmt.Errorf("no data found")
	}

	isMinor := false
	bpm := bpmMatch[1]
	key := keyMatch[1]
	if modeMatch[1] == "minor" {
		isMinor = true
	}

	if strings.Contains(key, "/") {
		parts := strings.Split(key, "/")
		key = parts[0]
	}

	key = strings.ReplaceAll(key, "♯", "#")
	key = strings.ReplaceAll(key, "♭", "b")

	if isMinor && !strings.HasSuffix(key, "m") {
		key += "m"
	}

	return &Metrics{
		BPM: bpm,
		Key: key,
	}, nil
}
