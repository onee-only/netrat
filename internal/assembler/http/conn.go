package http

import (
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/uuid"
	"github.com/onee-only/netrat/pkg/util"
)

type connInfo struct {
	id       uuid.UUID
	deadline time.Time
}

type connPairer struct {
	connections map[[2]gopacket.Flow]connInfo
	lock        sync.Mutex
}

func (p *connPairer) pairOrNew(dir [2]gopacket.Flow) (id uuid.UUID, found bool) {
	reversed := [2]gopacket.Flow{
		util.ReverseFlow(dir[0]),
		util.ReverseFlow(dir[1]),
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if info, ok := p.connections[reversed]; ok {
		if !info.deadline.Before(time.Now()) {
			delete(p.connections, reversed)
			return info.id, true
		}
	}

	id = uuid.New()
	info := connInfo{
		id:       id,
		deadline: time.Now().Add(timeout),
	}

	p.connections[dir] = info

	return id, false
}

func (p *connPairer) flush() (flushed int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	now := time.Now()
	for key, info := range p.connections {
		if info.deadline.Before(now) {
			delete(p.connections, key)
			flushed++
		}
	}

	return
}
