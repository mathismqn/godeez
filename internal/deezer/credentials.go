package deezer

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "godeez"
	keyringUser    = "default"
)

// Credentials is the JSON blob stored as a single system keyring secret. The
// password is kept alongside the ARL so an expired session can be renewed
// without prompting; see Login. Nothing here is ever written to disk by
// godeez itself.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ARL      string `json:"arl,omitempty"`
}

// loadCredentials returns the stored credentials, or nil with no error when
// the user has simply never logged in. That case is distinguished from a
// genuine keyring failure so callers can fall back to DEEZER_ARL instead of
// aborting.
func loadCredentials() (*Credentials, error) {
	secret, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("system keyring is unavailable: %v", err)
	}

	var creds Credentials
	if err := json.Unmarshal([]byte(secret), &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

func saveCredentials(creds *Credentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	if err := keyring.Set(keyringService, keyringUser, string(data)); err != nil {
		return fmt.Errorf("system keyring is unavailable: %v", err)
	}

	return nil
}

// ClearCredentials removes the stored credentials. Logging out when nothing
// is stored is not an error, so a missing entry is reported as success.
func ClearCredentials() error {
	if err := keyring.Delete(keyringService, keyringUser); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("system keyring is unavailable: %v", err)
	}

	return nil
}
