package deezer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// ErrInvalidARL reports that an ARL cookie was rejected. Callers should treat
// it as recoverable and re-login rather than as a hard failure; resolveARL
// relies on that distinction to decide whether to renew a stored session.
var ErrInvalidARL = errors.New("invalid or expired ARL cookie")

type Session struct {
	apiToken     string
	licenseToken string
	HTTPClient   *http.Client
	Premium      bool
}

// authenticate exchanges an ARL cookie for a Session. It returns
// ErrInvalidARL if the cookie is rejected.
//
// The endpoint answers 200 with an empty user for a bad cookie rather than an
// error status, so a zero user id is the only reliable signal that the ARL is
// no longer valid. A cookie jar is required because gw-light sets session
// cookies that later calls depend on.
//
// Premium is inferred from the offline listening options, which are the
// closest thing the payload carries to a subscription flag; it gates the
// higher quality formats.
func authenticate(ctx context.Context, arlCookie string) (*Session, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 20 * time.Second,
		Jar:     jar,
	}

	url := "https://www.deezer.com/ajax/gw-light.php?method=deezer.getUserData&input=3&api_version=1.0&api_token="
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return nil, ErrInvalidARL
	}

	opts := res.Results.User.Options
	return &Session{
		apiToken:     res.Results.APIToken,
		licenseToken: opts.LicenseToken,
		HTTPClient:   client,
		Premium:      opts.MobileOffline || opts.WebOffline,
	}, nil
}
