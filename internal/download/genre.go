package download

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

// Genres come from last.fm's community tags, which are free text and range
// from real genres to things like "seen live". These two lists are the filter
// that keeps only the useful ones. They are matched as substrings, so "deep
// house" is caught by "house".
//
// The split into two lists drives the ordering in filterTags, which prefers
// the electronic tag as the primary genre.
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
	// Escape the path segments: names containing '/', '?', or '#' would
	// otherwise change the URL structure and fetch the wrong page.
	reqURL := fmt.Sprintf("https://www.last.fm/music/%s/%s/+tags", url.PathEscape(artist), url.PathEscape(title))

	doc, err := fetchGenrePage(ctx, httpClient, reqURL)
	if err != nil {
		return "", err
	}

	// last.fm orders tags by popularity, so the first two are the consensus
	// view. Taking more starts pulling in mood and era tags that make a poor
	// genre field.
	tags := parseGenreTags(doc)
	if len(tags) > 2 {
		tags = tags[:2]
	}

	filtered := filterTags(tags)
	if len(filtered) == 0 {
		return "", errors.New("no data found")
	}

	return formatTags(filtered), nil
}

func fetchGenrePage(ctx context.Context, httpClient *http.Client, pageURL string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
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

// parseGenreTags reads the tag list out of a last.fm page. The selector
// tracks last.fm's current markup and is the first thing to break if they
// redesign; a failure here is non-fatal and simply leaves the genre unset.
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

// filterTags keeps only recognised genre tags, electronic ones first.
//
// It returns nothing at all unless at least one electronic tag matched, so a
// purely non-electronic track ends up with no genre rather than a partial
// one. Tags matching neither list are dropped.
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

// formatTags title cases the tags and joins them for the genre field.
// last.fm tags arrive in whatever case the tagger typed, so they are
// normalised rather than written through as is.
func formatTags(tags []string) string {
	formatted := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		words := strings.Fields(tag)
		for i, w := range words {
			r, size := utf8.DecodeRuneInString(w)
			words[i] = string(unicode.ToUpper(r)) + strings.ToLower(w[size:])
		}
		formatted = append(formatted, strings.Join(words, " "))
	}
	return strings.Join(formatted, " / ")
}
