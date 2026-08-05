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

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ARL      string `json:"arl,omitempty"`
}

func LoadCredentials() (*Credentials, error) {
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

func SaveCredentials(creds *Credentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	if err := keyring.Set(keyringService, keyringUser, string(data)); err != nil {
		return fmt.Errorf("system keyring is unavailable: %v", err)
	}

	return nil
}

func ClearCredentials() error {
	if err := keyring.Delete(keyringService, keyringUser); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("system keyring is unavailable: %v", err)
	}

	return nil
}
