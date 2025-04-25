package deezer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Song struct {
	ID           string `json:"SNG_ID"`
	Artist       string `json:"ART_NAME"`
	Title        string `json:"SNG_TITLE"`
	Version      string `json:"VERSION"`
	Cover        string `json:"ALB_PICTURE"`
	Contributors struct {
		MainArtists []string `json:"main_artist"`
		Composers   []string `json:"composer"`
		Authors     []string `json:"author"`
	} `json:"SNG_CONTRIBUTORS"`
	Duration    string `json:"DURATION"`
	Gain        string `json:"GAIN"`
	ISRC        string `json:"ISRC"`
	TrackNumber string `json:"TRACK_NUMBER"`
	TrackToken  string `json:"TRACK_TOKEN"`
}

func (s *Song) GetMediaData(licenseToken, quality string) (*Media, error) {
	var formats string

	switch quality {
	case "mp3_128":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	case "mp3_320":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"MP3_320"}]`
	case "flac":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"FLAC"}]`
	case "best":
		formats = `[{"cipher":"BF_CBC_STRIPE","format":"FLAC"},{"cipher":"BF_CBC_STRIPE","format":"MP3_320"},{"cipher":"BF_CBC_STRIPE","format":"MP3_128"}]`
	}

	reqBody := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":%s}],"track_tokens":["%s"]}`, licenseToken, formats, s.TrackToken)
	resp, err := http.Post("https://media.deezer.com/v1/get_url", "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var media Media
	err = json.Unmarshal(body, &media)
	if err != nil {
		return nil, err
	}

	if len(media.Errors) > 0 {
		if media.Errors[0].Code == 1000 {
			return nil, fmt.Errorf("invalid license token")
		}

		return nil, fmt.Errorf("%s", media.Errors[0].Message)
	}
	if len(media.Data) > 0 && len(media.Data[0].Errors) > 0 {
		if media.Data[0].Errors[0].Code == 2002 {
			return nil, fmt.Errorf("invalid track token")
		}

		return nil, fmt.Errorf("%s", media.Data[0].Errors[0].Message)
	}

	return &media, nil
}

func (s *Song) GetCoverImage() ([]byte, error) {
	url := fmt.Sprintf("https://e-cdn-images.dzcdn.net/images/cover/%s/500x500-000000-80-0-0.jpg", s.Cover)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (s *Song) GetTempoAndKey() (string, string, error) {
	client := &http.Client{}
	link, err := s.findSongLink(client)
	if err != nil {
		return "", "", err
	}

	html, err := fetchPage(client, link)
	if err != nil {
		return "", "", err
	}

	return parseBPMAndKey(html)
}

func (s *Song) findSongLink(client *http.Client) (string, error) {
	rootUrl := "https://songbpm.com"
	reqUrl := rootUrl + "/searches"

	values := url.Values{}
	values.Add("query", fmt.Sprintf("%s %s %s", s.Artist, s.Title, s.Version))

	req, err := http.NewRequest("POST", reqUrl, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://songbpm.com")

	resp, err := client.Do(req)
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

	var found bool
	var link string

	doc.Find("a.flex.flex-col").Each(func(i int, selection *goquery.Selection) {
		if strings.Contains(selection.Text(), s.Title) && strings.Contains(selection.Text(), s.Artist) {
			foundArtist := selection.Find("p.text-sm.font-light.uppercase").Text()
			foundTitle := selection.Find("p.pr-2.text-lg").Text()

			if strings.Contains(strings.ToLower(foundArtist), strings.ToLower(s.Artist)) && strings.Contains(strings.ToLower(foundTitle), strings.ToLower(s.Title)) {
				durationStr := strings.TrimSpace(selection.Find("div.flex-1.flex-col.items-center").Eq(1).Find("span.text-2xl").Text())
				parts := strings.Split(durationStr, ":")
				if len(parts) != 2 {
					return
				}
				minutes, err := strconv.Atoi(parts[0])
				if err != nil {
					fmt.Println(err)
					return
				}
				seconds, err := strconv.Atoi(parts[1])
				if err != nil {
					return
				}

				foundDuration := minutes*60 + seconds
				duration, err := strconv.Atoi(s.Duration)
				if err != nil {
					return
				}

				if foundDuration > (duration-2) || foundDuration < (duration+2) {
					link = selection.AttrOr("href", "")
					found = true

					return
				}
			}
		}
	})

	if !found {
		return "", fmt.Errorf("no data found")
	}

	return rootUrl + link, nil
}

func fetchPage(client *http.Client, link string) (string, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	return string(body), nil
}

func parseBPMAndKey(html string) (string, string, error) {
	bpmRegex := regexp.MustCompile(`tempo of <span[^>]*>(\d+) BPM`)
	bpmMatch := bpmRegex.FindStringSubmatch(html)

	keyRegex := regexp.MustCompile(`with a <span[^>]*>([A-G](?:♯|#|♭|b)?(?:/[A-G](?:♯|#|♭|b)?)?)</span> key`)
	keyMatch := keyRegex.FindStringSubmatch(html)

	modeRegex := regexp.MustCompile(`a  <span[^>]*>([a-z]+)</span> mode`)
	modeMatch := modeRegex.FindStringSubmatch(html)

	if len(bpmMatch) != 2 || len(keyMatch) != 2 || len(modeMatch) != 2 {
		return "", "", fmt.Errorf("no data found")
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

	return bpm, key, nil
}
