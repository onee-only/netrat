package manager

import (
	"github.com/google/uuid"
	"github.com/onee-only/netrat/internal/worker"
	"github.com/onee-only/netrat/pkg/stat"
	"github.com/pkg/errors"
)

type Manager struct {
	workers map[uuid.UUID]*worker.Worker
}

func New() *Manager {
	return &Manager{workers: make(map[uuid.UUID]*worker.Worker)}
}

func (m *Manager) RegisterWorker(w *worker.Worker) {
	m.workers[w.ID()] = w
}

func (m *Manager) All() (stats []stat.Worker) {
	stats = make([]stat.Worker, 0, len(m.workers))
	for _, w := range m.workers {
		stats = append(stats, w.ExportStats())
	}
	return
}

func (m *Manager) FetchStat(id uuid.UUID) (stat.Worker, error) {
	w, ok := m.workers[id]
	if !ok {
		return stat.Worker{}, errors.New("worker not found")
	}

	return w.ExportStats(), nil
}
