package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/mathismqn/godeez/internal/buildinfo"
	"github.com/mathismqn/godeez/internal/updater"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const updateNoticeAnnotation = "godeez:update-notice"

var RootCmd = &cobra.Command{
	Use:          "godeez",
	Short:        "GoDeez is a tool to download music from Deezer",
	SilenceUsage: true,
}

func Execute(ctx context.Context) error {
	var notice <-chan string
	if wantsUpdateNotice() {
		notice = updater.StartCheck(ctx)
	}

	err := RootCmd.ExecuteContext(ctx)

	printUpdateNotice(notice)

	return err
}

func wantsUpdateNotice() bool {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return false
	}

	args := os.Args[1:]
	if slices.Contains(args, "-h") || slices.Contains(args, "--help") {
		return false
	}

	target, _, err := RootCmd.Find(args)
	if err != nil || target == nil {
		return false
	}

	if target.Run == nil && target.RunE == nil {
		return false
	}

	for cmd := target; cmd != nil; cmd = cmd.Parent() {
		if cmd.Annotations[updateNoticeAnnotation] == "true" {
			return true
		}
	}

	return false
}

func printUpdateNotice(notice <-chan string) {
	select {
	case latest := <-notice:
		if latest == "" {
			return
		}

		fmt.Fprintf(os.Stderr, "\n  ┌ Update available: %s → %s\n  └ Run `godeez update` to install\n",
			buildinfo.Version(), latest)
	default:
	}
}
