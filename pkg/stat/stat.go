package stat

import (
	"time"

	"github.com/gofrs/uuid"
)

type Device struct {
	Name string

	NumWorkers int
}

type Worker struct {
	ID uuid.UUID

	CreatedAt time.Time
	Timeout   time.Duration
}
