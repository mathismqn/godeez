package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/mathismqn/godeez/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cmd.Execute(ctx); err != nil {
		stop()
		os.Exit(1)
	}
}
