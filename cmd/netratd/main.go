package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/onee-only/netrat/cmd/netratd/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), signalsToHandle...)
	defer stop()

	srv := server.New(server.Options{})

	if err := srv.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "netratd: %s\n", err)
		os.Exit(1)
	}
}
