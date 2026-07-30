package cmd

import (
	"fmt"

	"github.com/mathismqn/godeez/internal/auth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Deezer with your email and password",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.CheckGatewayEnv(); err != nil {
			return err
		}

		email, password, err := auth.PromptCredentials()
		if err != nil {
			return err
		}

		_, username, err := auth.Login(cmd.Context(), email, password)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully logged in as %s.\n", username)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
