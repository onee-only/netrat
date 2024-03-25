package server

import (
	"context"
	"errors"

	"github.com/onee-only/netrat/internal/msg"
)

type requestHandler func(ctx context.Context, r *msg.Request) (*msg.Response, error)

type actTable struct {
	lookup map[msg.RequestType]requestHandler
}

func (tbl *actTable) execute(ctx context.Context, req *msg.Request) (*msg.Response, error) {
	fn, ok := tbl.lookup[req.Type]
	if !ok {
		return nil, errors.New("request type not supported")
	}
	return fn(ctx, req)
}
