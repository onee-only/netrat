package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	goerrors "errors"

	"github.com/onee-only/netrat/cmd/netratd/server/internal/stat"
	"github.com/onee-only/netrat/cmd/netratd/server/internal/worker"
	"github.com/onee-only/netrat/internal/msg"
	"github.com/pkg/errors"
)

const ()

type Options struct {
	SocketAddr string
}

type Server struct {
	socketAddr string

	action *actTable

	stateStore *stat.Storage
	wManager   *worker.Manager
}

// New creates new netrat daemon.
func New(opts Options) *Server {
	srv := Server{
		socketAddr: opts.SocketAddr,
	}

	srv.action = &actTable{lookup: map[msg.RequestType]requestHandler{
		msg.RequestTypeDevList: func(ctx context.Context, r *msg.Request) (*msg.Response, error) {

			return nil, nil
		},
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
			errchan <- errors.Wrap(context.Cause(ctx), "server: waiting for connection")
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

		go func() {
			defer conn.Close()
			srv.accept(ctx, conn)
		}()
	}
}

func (srv *Server) accept(ctx context.Context, conn net.Conn) {
	var (
		lenBuf  = make([]byte, 4)
		recvBuf = make([]byte, 4096)
		buf     = new(bytes.Buffer)
	)

	for {
		req, err := srv.recv(ctx, conn, lenBuf, recvBuf, buf)
		if err != nil {
			res := msg.NewErrResponse(err)
			if err := srv.send(ctx, res, conn, lenBuf, buf); err != nil {
				// unexpected. fatal.
			}
			// TODO: maybe dump it?
			continue
		}

		res, err := srv.action.execute(ctx, req)
		if err != nil {
			res = msg.NewErrResponse(err)
			// TODO: maybe dump it?
		}

		if err := srv.send(ctx, res, conn, lenBuf, buf); err != nil {
			// unexpected. fatal.
		}
	}
}

func (srv *Server) recv(ctx context.Context, conn net.Conn, lenBuf, recvBuf []byte, buf *bytes.Buffer) (*msg.Request, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, context.Cause(ctx)
		default:
			conn.SetDeadline(time.Now().Add(time.Second))
		}

		_, err := conn.Read(lenBuf)
		if err != nil {
			if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
				continue
			}

			return nil, err
		}

		length := int(binary.LittleEndian.Uint32(lenBuf))

		buf.Grow(length)
		for length > 0 {
			select {
			case <-ctx.Done():
				return nil, context.Cause(ctx)
			default:
				conn.SetDeadline(time.Now().Add(time.Second))
			}

			n, err := conn.Read(recvBuf)
			if err != nil {
				if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
					continue
				}
				return nil, err
			}

			length -= n
			buf.Write(recvBuf[:n])
		}

		req, err := msg.DecodeRequest(buf)
		if err != nil {
			return nil, errors.Wrap(err, "server: decoding request")
		}

		return req, nil
	}
}

func (srv *Server) send(_ context.Context, res *msg.Response, conn net.Conn, lenBuf []byte, buf *bytes.Buffer) error {
	if err := res.Encode(buf); err != nil {
		return errors.Wrap(err, "server: encoding response")
	}

	binary.LittleEndian.PutUint32(lenBuf, uint32(buf.Len()))
	_, err := conn.Write(lenBuf)
	if err != nil {
		return errors.Wrap(err, "server: writing len data")
	}

	_, err = io.Copy(conn, buf)
	if err != nil {
		return errors.Wrap(err, "server: writing payload")
	}

	return nil
}
