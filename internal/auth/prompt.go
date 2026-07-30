package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func PromptCredentials() (string, string, error) {
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
