package deezer

import (
	"context"
	"errors"
	"fmt"
)

// resolveARL returns a usable ARL cookie from the stored credentials, logging
// in again if the stored one has expired.
//
// validate is supplied by the caller so the check can be the real
// authentication rather than a throwaway probe. Only ErrInvalidARL triggers a
// re-login: any other validation failure is most likely the network or Deezer
// being down, and silently re-sending the password in that case would turn a
// transient outage into a spurious login attempt.
func resolveARL(ctx context.Context, validate func(ctx context.Context, arl string) error) (string, error) {
	creds, err := loadCredentials()
	if err != nil {
		return "", err
	}

	if creds != nil && creds.ARL != "" {
		verr := validate(ctx, creds.ARL)
		if verr == nil {
			return creds.ARL, nil
		}

		if !errors.Is(verr, ErrInvalidARL) {
			return "", fmt.Errorf("stored session could not be validated: %w", verr)
		}
	}

	if creds != nil && creds.Email != "" && creds.Password != "" {
		arl, _, err := Login(ctx, creds.Email, creds.Password)
		if err != nil {
			return "", fmt.Errorf("stored session expired and could not be renewed automatically: %w", err)
		}

		return arl, nil
	}

	return "", errors.New("run 'godeez login' or export DEEZER_ARL environment variable")
}

// Login authenticates with email and password through the mobile gateway,
// persists the credentials to the system keyring, and returns the resulting
// ARL cookie and the account's display name.
//
// The password is stored, not just the ARL, because ARLs expire and renewing
// one without prompting the user again requires replaying the login.
func Login(ctx context.Context, email, password string) (string, string, error) {
	client, err := newMobileClient()
	if err != nil {
		return "", "", err
	}

	creds, username, err := client.login(ctx, email, password)
	if err != nil {
		return "", "", err
	}

	if err := saveCredentials(creds); err != nil {
		return "", "", err
	}

	return creds.ARL, username, nil
}
