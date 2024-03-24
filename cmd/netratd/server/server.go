package server

import (
	"bytes"
	"context"
	"net"
	"sync"
	"time"

	goerrors "errors"

	"github.com/pkg/errors"
)

const ()

type Options struct {
	SocketAddr string
}

type Server struct {
	socketAddr string

	connPool sync.Pool
}

// New creates new netrat daemon.
func New(opts Options) *Server {
	srv := Server{
		socketAddr: opts.SocketAddr,
	}

	srv.connPool = sync.Pool{New: func() any {
		return &handler{
			lenBuf:  make([]byte, 4),    // len(uint32)
			recvBuf: make([]byte, 4096), // 4KB
			buf:     new(bytes.Buffer),
		}
	}}

	return &srv
}

func (srv *Server) Run(ctx context.Context) (err error) {
	var wg sync.WaitGroup

	errchan := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		srv.serveUnix(ctx, errchan)
	}()

	wg.Wait()
	close(errchan)

	// drain any buffered errors.
	for e := range errchan {
		err = goerrors.Join(err, e)
	}

	return err
}

func (srv *Server) serveUnix(ctx context.Context, errchan chan<- error) {
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: srv.socketAddr})
	if err != nil {
		errchan <- errors.Wrap(err, "server: listening socket")
		return
	}
	defer listener.Close()

	for {
		select {
		case <-ctx.Done():
			errchan <- errors.Wrap(ctx.Err(), "server: waiting for connection")
			return
		default:
			listener.SetDeadline(time.Now().Add(time.Second))
		}

		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
				continue
			}
			errchan <- errors.Wrap(err, "server: accepting connection")
			return
		}

		h := srv.connPool.Get().(*handler)
		go func() {
			defer srv.connPool.Put(h)
			h.handle(ctx, conn)
		}()
	}
}
