package cmd

import (
	"fmt"

	"github.com/mathismqn/godeez/internal/deezer"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored Deezer credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := deezer.ClearCredentials(); err != nil {
				return err
			}

			fmt.Println("Successfully logged out.")

			return nil
		},
	}
}
