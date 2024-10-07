package song

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mathismqn/godeez/internal/models"
)

func GetMedia(s models.Song) (*models.MediaResponse, error) {
	licenseToken := ""
	reqBody := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":[{"cipher":"BF_CBC_STRIPE","format":"FLAC"}]}],"track_tokens":["%s"]}`, licenseToken, s.Token)

	resp, err := http.Post("https://media.deezer.com/v1/get_url", "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var media models.MediaResponse
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

	return &media, nil
}
