package stat

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/onee-only/netrat/pkg/stat"
)

type Storage struct {
	devs    map[string][]stat.Device
	devLock sync.Mutex

	workers     map[uuid.UUID]stat.Worker
	workersLock sync.Mutex
}

func NewStorage() *Storage {
	return &Storage{
		devs: make(map[string][]stat.Device),
	}
}
