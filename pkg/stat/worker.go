package stat

import (
	"time"

	"github.com/google/uuid"
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

	CreatedAt time.Time
	Timeout   time.Duration
}
