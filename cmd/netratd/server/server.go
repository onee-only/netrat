package server

import (
	"context"
	"net"
	"sync"
	"time"

	goerrors "errors"

	"github.com/pkg/errors"
)

const ()

type Options struct {
	Port uint16
}

type Server struct {
	socketAddr string

	connPool sync.Pool
}

// New creates new netrat daemon.
func New(opts Options) *Server {
	srv := Server{}

	return &srv
}

func (srv *Server) Run(ctx context.Context) (err error) {
	var wg sync.WaitGroup

	errchan := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()

		listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: srv.socketAddr})
		if err != nil {
			errchan <- errors.Wrap(err, "server: listening socket")
			return
		}
		defer listener.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				listener.SetDeadline(time.Now().Add(time.Second))
			}

			conn, err := listener.Accept()
			if err != nil {
				var netErr *net.OpError
				if errors.As(err, netErr) && netErr.Timeout() {
					continue
				}
				errchan <- errors.Wrap(err, "server: accepting connection")
				return
			}

			h := srv.connPool.Get().(*handler)
			go h.handle(ctx, conn)
		}
	}()

	wg.Wait()
	close(errchan)

	// drain any buffered errors.
	for e := range errchan {
		err = goerrors.Join(err, e)
	}

	return err
}
