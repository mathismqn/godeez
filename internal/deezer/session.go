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

type Session struct {
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

	var res struct {
		Results struct {
			APIToken string `json:"checkForm"`
			User     struct {
				ID      int `json:"USER_ID"`
				Options struct {
					LicenseToken  string `json:"license_token"`
					MobileOffline bool   `json:"mobile_offline"`
					WebOffline    bool   `json:"web_offline"`
				} `json:"OPTIONS"`
			} `json:"USER"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if res.Results.User.ID == 0 {
		return nil, fmt.Errorf("invalid arl cookie")
	}

	opts := res.Results.User.Options
	return &Session{
		APIToken:     res.Results.APIToken,
		LicenseToken: opts.LicenseToken,
		HttpClient:   client,
		Premium:      opts.MobileOffline || opts.WebOffline,
	}, nil
}
