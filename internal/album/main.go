package album

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/mathismqn/godeez/internal/models"
)

func GetDataByID(id string) (*models.Album, error) {
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

	var album models.Album
	err = json.Unmarshal([]byte(matches[1]), &album)
	if err != nil {
		return nil, err
	}

	return &album, nil
}
