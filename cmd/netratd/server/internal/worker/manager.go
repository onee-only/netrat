package worker

import (
	"github.com/gofrs/uuid"
)

type Manager struct {
	workers map[uuid.UUID]Worker
}

func (m *Manager) Register(worker *Worker) error {
	return nil
}
