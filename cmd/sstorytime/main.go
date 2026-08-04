package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Sole process root: signal-cancelable context for the whole CLI tree.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := Execute(ctx); err != nil {
		os.Exit(1)
	}
}
