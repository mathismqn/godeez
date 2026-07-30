package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mathismqn/godeez/internal/buildinfo"
	"github.com/mathismqn/godeez/internal/updater"
	"github.com/spf13/cobra"
)

var updateOpts struct {
	checkOnly bool
	force     bool
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update GoDeez to the latest version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runUpdate(cmd.Context())
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return err
	},
}

func runUpdate(ctx context.Context) error {
	if err := updater.CheckUpdatable(); err != nil {
		return err
	}

	u := updater.New()
	u.Out = os.Stdout
	current := buildinfo.Version()

	release, err := u.Latest(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latest := release.Version()
	fmt.Printf("Current: %s\nLatest:  %s\n", current, latest)

	if !updater.IsNewer(current, latest) && !updateOpts.force {
		fmt.Println("Already up to date.")

		return nil
	}

	if updateOpts.checkOnly {
		fmt.Printf("Run `godeez update` to install %s.\n", latest)

		return nil
	}

	if err := u.Apply(ctx, release); err != nil {
		return err
	}
	fmt.Printf("Updated to %s.\n", latest)

	return nil
}

func init() {
	RootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVar(&updateOpts.checkOnly, "check", false, "only report whether an update is available")
	updateCmd.Flags().BoolVar(&updateOpts.force, "force", false, "reinstall even if already up to date")
}
