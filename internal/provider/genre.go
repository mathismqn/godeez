package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var electronicKeywords = []string{
	"Ambient", "Bass", "Big Room", "Breakbeat", "Dance", "Disco", "Downtempo",
	"Drum And Bass", "Dub", "Dubstep", "EDM", "Electro", "Electronic", "Electronica",
	"Eurodance", "Gabber", "Garage", "Hardcore", "Hardstyle", "House", "Industrial",
	"Jungle", "Moombahton", "Synthpop", "Synthwave", "Techno", "Trance", "Trap",
	"Trip Hop", "Vaporwave",
}

var nonElectronicKeywords = []string{
	"Blues", "Chillout", "Classical", "Country", "Folk", "Funk", "Hip Hop", "Jazz",
	"Latin", "Metal", "Pop", "R&B", "Rap", "Reggae", "Rock", "Soul",
}

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

	if len(tags) > 2 {
		tags = tags[:2]
	}

	filteredTags := p.filterTags(tags)
	if len(filteredTags) == 0 {
		return "", fmt.Errorf("no data found")
	}

	genre := p.formatTags(filteredTags)
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

func (p GenreProvider) filterTags(tags []string) []string {
	var electronicTags []string
	var nonElectronicTags []string

	for _, tag := range tags {
		if p.isElectronicGenre(tag) {
			electronicTags = append(electronicTags, tag)
		} else if p.isNonElectronicGenre(tag) {
			nonElectronicTags = append(nonElectronicTags, tag)
		}
	}

	var filteredTags []string
	filteredTags = append(filteredTags, electronicTags...)

	if len(electronicTags) > 0 {
		filteredTags = append(filteredTags, nonElectronicTags...)
	}

	return filteredTags
}

func (p GenreProvider) isElectronicGenre(tag string) bool {
	tagLower := strings.ToLower(tag)
	for _, allowed := range electronicKeywords {
		if strings.Contains(tagLower, strings.ToLower(allowed)) {
			return true
		}
	}
	return false
}

func (p GenreProvider) isNonElectronicGenre(tag string) bool {
	tagLower := strings.ToLower(tag)
	for _, allowed := range nonElectronicKeywords {
		if strings.Contains(tagLower, strings.ToLower(allowed)) {
			return true
		}
	}
	return false
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
			if len(w) > 0 {
				words[i] = strings.ToUpper(string(w[0])) + strings.ToLower(w[1:])
			}
		}
		formatted = append(formatted, strings.Join(words, " "))
	}

	return strings.Join(formatted, " / ")
}
