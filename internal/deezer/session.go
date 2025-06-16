package deezer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
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
	HttpClient   *http.Client
	Premium      bool
}

func Authenticate(ctx context.Context, arlCookie string) (*Session, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 20 * time.Second,
		Jar:     jar,
	}

	url := "https://www.deezer.com/ajax/gw-light.php?method=deezer.getUserData&input=3&api_version=1.0&api_token="
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res UserDataResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if res.Results.User.Id == 0 {
		return nil, fmt.Errorf("invalid arl cookie")
	}

	isPremium := res.Results.User.Options.MobileOffline || res.Results.User.Options.WebOffline

	return &Session{
		ArlCookie:    arlCookie,
		APIToken:     res.Results.APIToken,
		LicenseToken: res.Results.User.Options.LicenseToken,
		HttpClient:   client,
		Premium:      isPremium,
	}, nil
}
