package auth

import (
	"context"
	"fmt"
)

func Resolve(ctx context.Context, validate func(ctx context.Context, arl string) error) (string, error) {
	creds, err := Load()
	if err != nil {
		return "", err
	}

	if creds != nil && creds.ARL != "" {
		if verr := validate(ctx, creds.ARL); verr == nil {
			return creds.ARL, nil
		}
	}

	if creds != nil && creds.Email != "" && creds.Password != "" {
		arl, _, err := Login(ctx, creds.Email, creds.Password)
		if err != nil {
			return "", fmt.Errorf("stored session expired and could not be renewed automatically: %w", err)
		}

		return arl, nil
	}

	return "", fmt.Errorf("run 'godeez login' or export DEEZER_ARL environment variable")
}

func Login(ctx context.Context, email, password string) (string, string, error) {
	client, err := newMobileClient()
	if err != nil {
		return "", "", err
	}

	creds, username, err := client.Login(ctx, email, password)
	if err != nil {
		return "", "", err
	}

	if err := Save(creds); err != nil {
		return "", "", err
	}

	return creds.ARL, username, nil
}
