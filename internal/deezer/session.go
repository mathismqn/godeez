package deezer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
)

type UserDataResponse struct {
	Results struct {
		APIToken string `json:"checkForm"`
		User     struct {
			Id      int `json:"USER_ID"`
			Options struct {
				LicenseToken  string `json:"license_token"`
				MobileOffline bool   `json:"mobile_offline"`
				WebOffline    bool   `json:"web_offline"`
			} `json:"OPTIONS"`
		} `json:"USER"`
	} `json:"results"`
}

type Session struct {
	ArlCookie    string
	APIToken     string
	LicenseToken string
	Client       *http.Client
}

func Authenticate(arlCookie string) (*Session, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar: jar,
	}

	url := "https://www.deezer.com/ajax/gw-light.php?method=deezer.getUserData&input=3&api_version=1.0&api_token="
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "arl",
		Value: arlCookie,
	})

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var res UserDataResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if res.Results.User.Id == 0 {
		return nil, fmt.Errorf("invalid arl cookie")
	}
	if !res.Results.User.Options.MobileOffline && !res.Results.User.Options.WebOffline {
		return nil, fmt.Errorf("premium account required")
	}

	return &Session{
		ArlCookie:    arlCookie,
		APIToken:     res.Results.APIToken,
		LicenseToken: res.Results.User.Options.LicenseToken,
		Client:       client,
	}, nil
}
