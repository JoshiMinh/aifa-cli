package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"aifiler/internal/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	exitCode := cli.NewApp().Run(ctx, os.Args[1:])
	os.Exit(exitCode)
}
