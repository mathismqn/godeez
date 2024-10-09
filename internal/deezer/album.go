package deezer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type Album struct {
	Data struct {
		ID          string `json:"ALB_ID"`
		Name        string `json:"ALB_TITLE"`
		ArtistID    string `json:"ART_ID"`
		ArtistName  string `json:"ART_NAME"`
		CoverID     string `json:"ALB_PICTURE"`
		ReleaseData string `json:"PHYSICAL_RELEASE_DATE"`
	} `json:"DATA"`
	Songs struct {
		Data []Song `json:"data"`
	} `json:"SONGS"`
}

func GetAlbumData(id string) (*Album, error) {
	url := fmt.Sprintf("https://www.deezer.com/en/album/%s", id)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`window\.__DZR_APP_STATE__ = (\{.*\})`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		return nil, fmt.Errorf("error parsing response")
	}

	var album Album
	err = json.Unmarshal([]byte(matches[1]), &album)
	if err != nil {
		return nil, err
	}

	return &album, nil
}