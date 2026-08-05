package deezer

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	gatewayBaseURL   = "https://api.deezer.com/1.0/gateway.php"
	gatewayUserAgent = "Deezer/6.1.22.49 (Android; 9; Tablet; us) innotek GmbH VirtualBox"
	nonceAlphabet    = "012345689abdef"
)

const (
	deviceOS       = "Android"
	deviceName     = "VirtualBox"
	deviceType     = "tablet"
	deviceModel    = "VirtualBox"
	devicePlatform = "innotek GmbH_x86_64_9"
	deviceSerial   = ""
)

type mobileClient struct {
	httpClient *http.Client
	apiKey     string
	gwKey      []byte
	sid        string
}

func CheckGatewayEnv() error {
	_, _, err := gatewayEnv()
	return err
}

func gatewayEnv() (string, string, error) {
	apiKey := os.Getenv("DEEZER_MOBILE_API_KEY")
	gwKey := os.Getenv("DEEZER_MOBILE_GW_KEY")
	if apiKey == "" || gwKey == "" {
		return "", "", errors.New("DEEZER_MOBILE_API_KEY and DEEZER_MOBILE_GW_KEY must be set to use email/password login")
	}

	if len(gwKey) != aes.BlockSize {
		return "", "", fmt.Errorf("DEEZER_MOBILE_GW_KEY must be exactly %d bytes long", aes.BlockSize)
	}

	return apiKey, gwKey, nil
}

func newMobileClient() (*mobileClient, error) {
	apiKey, gwKey, err := gatewayEnv()
	if err != nil {
		return nil, err
	}

	return &mobileClient{
		httpClient: &http.Client{Timeout: 20 * time.Second},
		apiKey:     apiKey,
		gwKey:      []byte(gwKey),
	}, nil
}

func (m *mobileClient) login(ctx context.Context, email, password string) (*Credentials, string, error) {
	token, tokenKey, userKey, err := m.authenticate(ctx)
	if err != nil {
		return nil, "", err
	}

	if err := m.checkToken(ctx, token, tokenKey); err != nil {
		return nil, "", err
	}

	arl, username, err := m.userAuth(ctx, email, password, userKey)
	if err != nil {
		return nil, "", err
	}

	return &Credentials{Email: email, Password: password, ARL: arl}, username, nil
}

func (m *mobileClient) authenticate(ctx context.Context) (string, string, string, error) {
	body, err := m.gatewayRequest(ctx, "mobile_auth", http.MethodGet, "uniq_id", genUniqID(), nil)
	if err != nil {
		return "", "", "", err
	}

	var res struct {
		Results struct {
			Token string `json:"TOKEN"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", "", "", err
	}

	if strings.Contains(string(body), "Undefined or invalid API key") {
		return "", "", "", errors.New("DEEZER_MOBILE_API_KEY is invalid")
	}
	if strings.Contains(string(body), "GATEWAY_ERROR") || res.Results.Token == "" {
		return "", "", "", errors.New("unexpected response from gateway")
	}

	encrypted, err := hex.DecodeString(res.Results.Token)
	if err != nil {
		return "", "", "", err
	}

	decrypted, err := ecbDecrypt(m.gwKey, encrypted)
	if err != nil {
		return "", "", "", err
	}

	if len(decrypted) < 96 {
		return "", "", "", errors.New("unexpected response from gateway")
	}

	token := string(decrypted[0:64])
	tokenKey := string(decrypted[64:80])
	userKey := string(decrypted[80:96])

	return token, tokenKey, userKey, nil
}

func (m *mobileClient) checkToken(ctx context.Context, token, tokenKey string) error {
	encrypted, err := ecbEncrypt([]byte(tokenKey), []byte(token))
	if err != nil {
		return err
	}
	authToken := hex.EncodeToString(encrypted)

	body, err := m.gatewayRequest(ctx, "api_checkToken", http.MethodGet, "auth_token", authToken, nil)
	if err != nil {
		return err
	}

	var res struct {
		Results string `json:"results"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if res.Results == "" {
		return errors.New("unexpected response from gateway")
	}
	m.sid = res.Results

	return nil
}

func (m *mobileClient) userAuth(ctx context.Context, email, password, userKey string) (string, string, error) {
	encryptedPassword, err := ecbEncrypt([]byte(userKey), zeroPad([]byte(password)))
	if err != nil {
		return "", "", err
	}

	payload := map[string]string{
		"mail":                              email,
		"password":                          hex.EncodeToString(encryptedPassword),
		"device_serial":                     deviceSerial,
		"platform":                          devicePlatform,
		"custo_version_id":                  "",
		"custo_partner":                     "",
		"model":                             deviceModel,
		"device_name":                       deviceName,
		"device_os":                         deviceOS,
		"device_type":                       deviceType,
		"google_play_services_availability": "1",
		"consent_string":                    "",
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}

	body, err := m.gatewayRequest(ctx, "mobile_userAuth", http.MethodPost, "", "", jsonBody)
	if err != nil {
		return "", "", err
	}

	if strings.Contains(string(body), "USER_AUTH_ERROR") {
		return "", "", errors.New("invalid email or password")
	}

	var res struct {
		Results struct {
			ARL      string `json:"ARL"`
			UserID   int    `json:"USER_ID"`
			BlogName string `json:"BLOG_NAME"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", "", err
	}

	if res.Results.ARL == "" || res.Results.UserID == 0 {
		return "", "", errors.New("unexpected response from gateway")
	}

	return res.Results.ARL, res.Results.BlogName, nil
}

func (m *mobileClient) gatewayRequest(ctx context.Context, method, httpMethod, paramKey, paramValue string, jsonBody []byte) ([]byte, error) {
	u, err := url.Parse(gatewayBaseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("method", method)
	q.Set("api_key", m.apiKey)
	q.Set("output", "3")
	if httpMethod == http.MethodPost {
		q.Set("input", "3")
	}
	if m.sid != "" {
		q.Set("sid", m.sid)
	}
	if paramKey != "" {
		q.Set(paramKey, paramValue)
	}
	u.RawQuery = q.Encode()

	var reqBody io.Reader
	if jsonBody != nil {
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, u.String(), reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", gatewayUserAgent)
	if jsonBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func genUniqID() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = nonceAlphabet[rand.IntN(len(nonceAlphabet))]
	}

	return string(b)
}
