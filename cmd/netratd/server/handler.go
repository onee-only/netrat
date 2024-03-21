package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"time"

	"github.com/pkg/errors"
)

type handler struct {
	lenBuf, recvBuf []byte
	buf             *bytes.Buffer
}

// It will eventually close connection
func (h *handler) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	defer h.buf.Reset()

	for {
		err := h.recv(ctx, conn)
		if err != nil {
			// TODO: handle error and dump it.
			return
		}

		// do something.
	}
}

func (h *handler) recv(ctx context.Context, conn net.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn.SetDeadline(time.Now().Add(time.Second))
		}

		_, err := conn.Read(h.lenBuf)
		if err != nil {
			var netErr *net.OpError
			if errors.As(err, netErr) && netErr.Timeout() {
				continue
			}
			return err
		}

		length := int(binary.LittleEndian.Uint32(h.lenBuf))

		h.buf.Grow(length)
		for length > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				conn.SetDeadline(time.Now().Add(time.Second))
			}

			n, err := conn.Read(h.recvBuf)
			if err != nil {
				var netErr *net.OpError
				if errors.As(err, netErr) && netErr.Timeout() {
					continue
				}
				return err
			}

			length -= n
			h.buf.Write(h.recvBuf[:n])
		}

		return nil
	}
}
