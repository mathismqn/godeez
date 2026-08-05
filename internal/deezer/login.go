package deezer

import (
	"context"
	"errors"
	"fmt"
)

func resolveARL(ctx context.Context, validate func(ctx context.Context, arl string) error) (string, error) {
	creds, err := LoadCredentials()
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

func Login(ctx context.Context, email, password string) (string, string, error) {
	client, err := newMobileClient()
	if err != nil {
		return "", "", err
	}

	creds, username, err := client.login(ctx, email, password)
	if err != nil {
		return "", "", err
	}

	if err := SaveCredentials(creds); err != nil {
		return "", "", err
	}

	return creds.ARL, username, nil
}
