package worker

import (
	"github.com/google/gopacket/pcap"
	"github.com/google/uuid"
)

type ListenOptions struct {
}

type Worker struct {
	ID uuid.UUID

	handle *pcap.Handle
	cancel func()
}

func New(opts *ListenOptions) *Worker {
	w := &Worker{}
}
