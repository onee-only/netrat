package msg

import (
	"encoding/gob"
	"errors"
	"io"

	"github.com/google/uuid"
)

type Response struct {
	Payload any
	ErrMsg  string
}

func (r *Response) Err() error {
	if r.ErrMsg == "" {
		return nil
	}
	return errors.New(r.ErrMsg)
}

func (r *Response) Encode(w io.Writer) error {
	return gob.NewEncoder(w).Encode(r)
}

func DecodeResponse(r io.Reader) (res *Response, err error) {
	res = &Response{}
	err = gob.NewDecoder(r).Decode(res)
	return
}

func NewErrResponse(err error) *Response {
	return &Response{
		Payload: EmptyPayload{},
		ErrMsg:  err.Error(),
	}
}

type WorkerIDPayload struct {
	ID uuid.UUID
}

func registerResponse() {
	gob.Register(WorkerIDPayload{})
}
