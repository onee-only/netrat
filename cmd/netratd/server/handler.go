package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/onee-only/netrat/internal/msg"
	"github.com/pkg/errors"
)

type handler struct {
	lenBuf, recvBuf []byte

	buf    *bytes.Buffer
	action *actTable
}

// It will eventually close connection
func (h *handler) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	defer h.buf.Reset()

	for {
		req, err := h.recv(ctx, conn)
		if err != nil {
			res := msg.NewErrResponse(err)
			if err := h.send(ctx, res, conn); err != nil {
				// unexpected. fatal.
			}
			// TODO: maybe dump it?
			continue
		}

		res, err := h.action.execute(ctx, req)
		if err != nil {
			res = msg.NewErrResponse(err)
			// TODO: maybe dump it?
		}

		if err := h.send(ctx, res, conn); err != nil {
			// unexpected. fatal.
		}
	}
}

func (h *handler) recv(ctx context.Context, conn net.Conn) (*msg.Request, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			conn.SetDeadline(time.Now().Add(time.Second))
		}

		_, err := conn.Read(h.lenBuf)
		if err != nil {
			if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
				continue
			}

			return nil, err
		}

		length := int(binary.LittleEndian.Uint32(h.lenBuf))

		h.buf.Grow(length)
		for length > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				conn.SetDeadline(time.Now().Add(time.Second))
			}

			n, err := conn.Read(h.recvBuf)
			if err != nil {
				if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
					continue
				}
				return nil, err
			}

			length -= n
			h.buf.Write(h.recvBuf[:n])
		}

		req, err := msg.DecodeRequest(h.buf)
		if err != nil {
			return nil, errors.Wrap(err, "server: decoding request")
		}

		return req, nil
	}
}

func (h *handler) send(_ context.Context, res *msg.Response, conn net.Conn) error {
	if err := res.Encode(h.buf); err != nil {
		return errors.Wrap(err, "server: encoding response")
	}

	binary.LittleEndian.PutUint32(h.lenBuf, uint32(h.buf.Len()))
	_, err := conn.Write(h.lenBuf)
	if err != nil {
		return errors.Wrap(err, "server: writing len data")
	}

	_, err = io.Copy(conn, h.buf)
	if err != nil {
		return errors.Wrap(err, "server: writing payload")
	}

	return nil
}
