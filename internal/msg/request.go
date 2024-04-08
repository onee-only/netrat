package msg

import (
	"encoding/gob"
	"io"

	"github.com/onee-only/netrat/internal/worker"
)

type RequestType uint8

const (
	RequestTypeListen RequestType = 1 + iota
	RequestTypeWorkerList
	RequestTypeWorkerStat
)

type Request struct {
	Type RequestType

	Payload any
}

func (r *Request) Encode(w io.Writer) error {
	return gob.NewEncoder(w).Encode(r)
}

func DecodeRequest(r io.Reader) (req *Request, err error) {
	req = &Request{}
	err = gob.NewDecoder(r).Decode(req)
	return
}

type WorkerInitPayload struct {
	Opts worker.WorkerOptions
}

func registerRequest() {
	gob.Register(WorkerInitPayload{})
}