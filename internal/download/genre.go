package download

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var electronicKeywords = toLower([]string{
	"Ambient", "Bass", "Big Room", "Breakbeat", "Dance", "Disco", "Downtempo",
	"Drum And Bass", "Dub", "Dubstep", "EDM", "Electro", "Electronic", "Electronica",
	"Eurodance", "Gabber", "Garage", "Hardcore", "Hardstyle", "House", "Industrial",
	"Jungle", "Moombahton", "Synthpop", "Synthwave", "Techno", "Trance", "Trap",
	"Trip Hop", "Vaporwave",
})

var nonElectronicKeywords = toLower([]string{
	"Blues", "Chillout", "Classical", "Country", "Folk", "Funk", "Hip Hop", "Jazz",
	"Latin", "Metal", "Pop", "R&B", "Rap", "Reggae", "Rock", "Soul",
})

func toLower(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = strings.ToLower(s)
	}
	return out
}

func fetchGenre(ctx context.Context, httpClient *http.Client, artist, title string) (string, error) {
	reqURL := fmt.Sprintf("https://www.last.fm/music/%s/%s/+tags", artist, title)

	doc, err := fetchGenrePage(ctx, httpClient, reqURL)
	if err != nil {
		return "", err
	}

	tags := parseGenreTags(doc)
	if len(tags) > 2 {
		tags = tags[:2]
	}

	filtered := filterTags(tags)
	if len(filtered) == 0 {
		return "", fmt.Errorf("no data found")
	}

	return formatTags(filtered), nil
}

func fetchGenrePage(ctx context.Context, httpClient *http.Client, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	return goquery.NewDocumentFromReader(resp.Body)
}

func parseGenreTags(doc *goquery.Document) []string {
	var tags []string
	doc.Find("ol.big-tags .big-tags-item-name a").Each(func(_ int, s *goquery.Selection) {
		if tag := strings.TrimSpace(s.Text()); tag != "" {
			tags = append(tags, tag)
		}
	})
	return tags
}

func matchesKeyword(tag string, keywords []string) bool {
	tagLower := strings.ToLower(tag)
	for _, kw := range keywords {
		if strings.Contains(tagLower, kw) {
			return true
		}
	}
	return false
}

func filterTags(tags []string) []string {
	var electronic, nonElectronic []string

	for _, tag := range tags {
		if matchesKeyword(tag, electronicKeywords) {
			electronic = append(electronic, tag)
		} else if matchesKeyword(tag, nonElectronicKeywords) {
			nonElectronic = append(nonElectronic, tag)
		}
	}

	if len(electronic) > 0 {
		return append(electronic, nonElectronic...)
	}
	return electronic
}

func formatTags(tags []string) string {
	formatted := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		words := strings.Fields(tag)
		for i, w := range words {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
		formatted = append(formatted, strings.Join(words, " "))
	}
	return strings.Join(formatted, " / ")
}
