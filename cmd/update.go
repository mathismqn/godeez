package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mathismqn/godeez/internal/buildinfo"
	"github.com/mathismqn/godeez/internal/update"
	"github.com/spf13/cobra"
)

type updateOptions struct {
	checkOnly bool
	force     bool
}

func newUpdateCmd() *cobra.Command {
	opts := &updateOptions{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update GoDeez to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runUpdate(cmd.Context(), opts)
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		},
	}

	cmd.Flags().BoolVar(&opts.checkOnly, "check", false, "only report whether an update is available")
	cmd.Flags().BoolVar(&opts.force, "force", false, "reinstall even if already up to date")

	return cmd
}

func runUpdate(ctx context.Context, opts *updateOptions) error {
	if err := update.CheckUpdatable(); err != nil {
		return err
	}

	u := update.New()
	u.Out = os.Stdout
	current := buildinfo.Version()

	release, err := u.Latest(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latest := release.Version()
	fmt.Printf("Current: %s\nLatest:  %s\n", current, latest)

	if !update.IsNewer(current, latest) && !opts.force {
		fmt.Println("Already up to date.")

		return nil
	}

	if opts.checkOnly {
		fmt.Printf("Run `godeez update` to install %s.\n", latest)

		return nil
	}

	if err := u.Apply(ctx, release); err != nil {
		return err
	}
	fmt.Printf("Updated to %s.\n", latest)

	return nil
}
