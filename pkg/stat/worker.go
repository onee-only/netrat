package stat

import (
	"time"

	"github.com/google/gopacket"
	"github.com/google/uuid"
	"github.com/onee-only/netrat/pkg/assemble"
)

type WorkerState uint8

const (
	WorkerStateInit WorkerState = 1 + iota
	WorkerStateUp
	WorkerStateFin
	WorkerStateCancel
)

type Worker struct {
	ID uuid.UUID

	Live bool
	Src  string

	SnapLen     int32
	Promiscuous bool
	BPFFilter   string

	Captures  []gopacket.LayerType
	Assembles []assemble.AssembleType

	State     WorkerState
	CreatedAt time.Time
	Timeout   time.Duration
}
