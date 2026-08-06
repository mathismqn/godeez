// Package cmd defines the godeez command line: the root command and its
// download, login, logout, update and version subcommands. It is a thin layer
// that parses flags and delegates to the internal packages.
package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/mathismqn/godeez/internal/buildinfo"
	"github.com/mathismqn/godeez/internal/update"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// updateNoticeAnnotation marks the commands that may print an update notice.
// It is an annotation rather than a field because it is inherited: marking
// the download command opts in all of its subcommands.
const updateNoticeAnnotation = "godeez:update-notice"

// Execute runs the CLI.
//
// The update check is started before the command and collected after it, so
// the network round trip overlaps with work the user actually asked for
// instead of adding to the startup time.
func Execute(ctx context.Context) error {
	root := newRootCmd()

	var notice <-chan string
	if wantsUpdateNotice(root) {
		notice = update.StartCheck(ctx)
	}

	err := root.ExecuteContext(ctx)

	printUpdateNotice(notice)

	return err
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "godeez",
		Short:        "GoDeez is a tool to download music from Deezer",
		SilenceUsage: true,
	}

	root.AddCommand(
		newDownloadCmd(),
		newLoginCmd(),
		newLogoutCmd(),
		newUpdateCmd(),
		newVersionCmd(),
	)

	return root
}

// wantsUpdateNotice decides whether this invocation should check for updates.
//
// The aim is to nag only during real interactive use. Notices are suppressed
// when stderr is not a terminal, so they cannot corrupt piped or scripted
// output; on help output, where they are noise; on commands that only print
// their usage; and on anything not explicitly opted in via the annotation,
// which notably keeps `godeez version` and `godeez update` quiet.
func wantsUpdateNotice(root *cobra.Command) bool {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return false
	}

	args := os.Args[1:]
	if slices.Contains(args, "-h") || slices.Contains(args, "--help") {
		return false
	}

	target, _, err := root.Find(args)
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

// printUpdateNotice prints the notice only if the check has already finished.
//
// The non-blocking receive is the point: the command is done and the user
// should get their prompt back, so a check that is still in flight is
// dropped rather than waited on. A nil channel, meaning no check was started,
// takes the same path.
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
