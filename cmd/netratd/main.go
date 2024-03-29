package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/onee-only/netrat/cmd/netratd/server"
	"github.com/onee-only/netrat/internal/config"
)

func init() {
	maxprocs.Set()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), signalsToHandle...)
	defer stop()

	opts := server.Options{
		SocketAddr: config.DefaultServerAddr,
	}

	srv := server.New(opts)

	if err := srv.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "netratd: %s\n", err)
		os.Exit(1)
	}
}
