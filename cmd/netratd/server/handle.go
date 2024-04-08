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

func (srv *Server) HandleList(ctx context.Context, r *msg.Request) (*msg.Response, error) {
	workers := srv.workManager.All()

	return &msg.Response{
		Payload: msg.WorkerListPayload{Workers: workers},
	}, nil
}

func (srv *Server) HandleStat(ctx context.Context, r *msg.Request) (*msg.Response, error) {
	p := r.Payload.(msg.WorkerIDPayload)
	stat, err := srv.workManager.FetchStat(p.ID)
	if err != nil {
		return nil, err
	}

	return &msg.Response{
		Payload: msg.WorkerStatPayload{
			Stat: stat,
		},
	}, nil
}
