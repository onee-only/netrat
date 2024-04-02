package http

import (
	"context"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/onee-only/netrat/internal/assembler"
	"github.com/onee-only/netrat/internal/container"
	"github.com/onee-only/netrat/internal/storage"
)

const timeout = 30 * time.Second

type HTTPAssembler struct {
	tcpasm     *tcpassembly.Assembler
	connPairer *connPairer
	storage    storage.AssembleObjectStorage

	cancel func()
}

var _ assembler.Assembler = (*HTTPAssembler)(nil)

func NewHTTPAssembler(s storage.AssembleObjectStorage) *HTTPAssembler {
	asm := &HTTPAssembler{
		storage:    s,
		connPairer: &connPairer{connections: make(map[[2]gopacket.Flow]connInfo)},
	}

	ctx, cancel := context.WithCancel(context.Background())

	asm.cancel = cancel

	factory := &httpStreamFactory{
		pctx:       ctx,
		asmStorage: asm.storage,
		connPairer: asm.connPairer,
	}

	streamPool := tcpassembly.NewStreamPool(factory)
	asm.tcpasm = tcpassembly.NewAssembler(streamPool)

	go func() {
		t := time.NewTicker(timeout)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				asm.tcpasm.FlushOlderThan(time.Now().Add(-timeout))
				asm.connPairer.flush()
			}
		}
	}()

	return asm
}

func (asm *HTTPAssembler) Provide(packet container.Packet) {
	tcpPacket := packet.TransportLayer().(*layers.TCP)

	asm.tcpasm.AssembleWithTimestamp(
		packet.NetworkLayer().NetworkFlow(),
		tcpPacket, packet.Metadata().Timestamp,
	)
}

func (asm *HTTPAssembler) Cancel() {
	asm.cancel()
}
