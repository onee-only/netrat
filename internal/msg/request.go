package msg

import (
	"encoding/gob"
	"io"
)

type RequestType uint8

const (
	RequestTypeDevList RequestType = 1 + iota
)

type Request struct {
	Type RequestType

	Payload any
}

func DecodeRequest(r io.Reader) (req *Request, err error) {
	req = &Request{}
	err = gob.NewDecoder(r).Decode(req)
	return
}
