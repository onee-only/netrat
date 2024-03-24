package msg

import (
	"encoding/gob"
	"errors"
	"io"
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

func NewErrResponse(err error) *Response {
	return &Response{
		Payload: EmptyPayload{},
		ErrMsg:  err.Error(),
	}
}
