// Command godeez downloads music from Deezer. See the cmd package for the
// command line surface and the internal packages for the download pipeline.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/mathismqn/godeez/cmd"
)

func main() {
	// The interrupt-cancelled context is threaded through every network call
	// and file write, so Ctrl-C unwinds the download cleanly and leaves no
	// partial files behind rather than killing the process mid-write.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Nothing is printed here because cobra has already reported the error.
	// stop is called explicitly since the deferred call would not run before
	// os.Exit.
	if err := cmd.Execute(ctx); err != nil {
		stop()
		os.Exit(1)
	}
}
