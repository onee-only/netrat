package server

import (
	"context"
	"log"

	"github.com/onee-only/netrat/internal/msg"
	"github.com/onee-only/netrat/internal/worker"
)

func (srv *Server) HandleListen(ctx context.Context, r *msg.Request) (*msg.Response, error) {
	p := r.Payload.(msg.WorkerInitPayload)
	w, ctx, err := worker.New(ctx, &p.Opts)
	if err != nil {
		return nil, err
	}

	srv.workManager.RegisterWorker(w)

	go func() {
		if err := w.Exec(ctx); err != nil {
			log.Println(err)
		}
	}()

	return &msg.Response{
		Payload: msg.WorkerIDPayload{ID: w.ID()},
	}, nil
}
