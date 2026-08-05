package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log in to Deezer with your email and password",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := deezer.CheckGatewayEnv(); err != nil {
				return err
			}

			err := runLogin(cmd.Context())
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		},
	}
}

func runLogin(ctx context.Context) error {
	email, password, err := promptCredentials(ctx)
	if err != nil {
		return err
	}

	_, username, err := deezer.Login(ctx, email, password)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully logged in as %s.\n", username)

	return nil
}

func promptCredentials(ctx context.Context) (string, string, error) {
	oldState, stateErr := term.GetState(int(os.Stdin.Fd()))

	type credentials struct {
		email    string
		password string
		err      error
	}
	resultChan := make(chan credentials, 1)
	go func() {
		var c credentials
		c.email, c.password, c.err = readCredentials()
		resultChan <- c
	}()

	select {
	case c := <-resultChan:
		return c.email, c.password, c.err
	case <-ctx.Done():
		if stateErr == nil {
			term.Restore(int(os.Stdin.Fd()), oldState)
		}
		fmt.Println()

		return "", "", ctx.Err()
	}
}

func readCredentials() (string, string, error) {
	fmt.Print("Email: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", "", err
	}
	email := strings.TrimSpace(line)

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", "", err
	}

	return email, string(passwordBytes), nil
}
