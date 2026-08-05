package cmd

import (
	"bufio"
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

			email, password, err := promptCredentials()
			if err != nil {
				return err
			}

			_, username, err := deezer.Login(cmd.Context(), email, password)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully logged in as %s.\n", username)

			return nil
		},
	}
}

func promptCredentials() (string, string, error) {
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
