package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gobkd/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		panic(err)
	}
}
