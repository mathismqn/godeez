package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type GenreProvider struct{}

func (p GenreProvider) Fetch(ctx context.Context, httpClient *http.Client, artist, title string) (string, error) {
	reqUrl := fmt.Sprintf("https://www.last.fm/music/%s/%s/+tags", artist, title)
	doc, err := p.fetchPage(ctx, httpClient, reqUrl)
	if err != nil {
		return "", err
	}

	tags := p.parse(doc)
	if len(tags) == 0 {
		return "", fmt.Errorf("no data found")
	}

	if len(tags) > 1 {
		tags = tags[:2]
	}
	genre := p.formatTags(tags)

	return genre, nil
}

func (p GenreProvider) fetchPage(ctx context.Context, httpClient *http.Client, reqUrl string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (p GenreProvider) parse(doc *goquery.Document) []string {
	var tags []string
	doc.Find("ol.big-tags .big-tags-item-name a").Each(func(_ int, s *goquery.Selection) {
		tag := strings.TrimSpace(s.Text())
		if tag != "" {
			tags = append(tags, tag)
		}
	})

	return tags
}

func (p GenreProvider) formatTags(tags []string) string {
	var formatted []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		words := strings.Fields(tag)
		for i, w := range words {
			words[i] = strings.Title(w)
		}
		formatted = append(formatted, strings.Join(words, " "))
	}

	return strings.Join(formatted, "/")
}
