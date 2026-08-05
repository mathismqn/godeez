package cmd

import (
	"fmt"
	"runtime"

	"github.com/mathismqn/godeez/internal/buildinfo"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current version of GoDeez",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("godeez %s\n", buildinfo.Version())

			if commit := buildinfo.Commit(); commit != "" {
				fmt.Printf("  commit:   %s\n", commit)
			}
			if date := buildinfo.Date(); date != "" {
				fmt.Printf("  built:    %s\n", date)
			}

			fmt.Printf("  go:       %s\n", runtime.Version())
			fmt.Printf("  platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
