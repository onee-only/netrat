package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/google/gopacket/layers"
	"github.com/google/uuid"
	"github.com/onee-only/netrat/internal/assembler"
	asmfactory "github.com/onee-only/netrat/internal/assembler/factory"
	"github.com/onee-only/netrat/internal/config"
	"github.com/onee-only/netrat/internal/container"
	"github.com/onee-only/netrat/internal/storage"
	storagefactory "github.com/onee-only/netrat/internal/storage/packet/factory"
	"github.com/onee-only/netrat/pkg/assemble"
	"github.com/onee-only/netrat/pkg/stat"
	"github.com/pkg/errors"
)

type WorkerOptions struct {
	ListenOptions

	AssembleTypes []assemble.AssembleType
}

func (o *WorkerOptions) Validate() (*WorkerOptions, error) {
	if o == nil {
		o = &WorkerOptions{}
	}

	if len(o.AssembleTypes) > 0 && !slices.Contains(o.CaptureLayers, layers.LayerTypeTCP) {
		return nil, errors.New("listener: cannot use assembler when capture layer is not tcp")
	}

	invalidAsmTypes := make([]assemble.AssembleType, 0)
	for _, t := range o.AssembleTypes {
		if !t.Valid() {
			invalidAsmTypes = append(invalidAsmTypes, t)
		}
	}

	if len(invalidAsmTypes) > 0 {
		return nil, fmt.Errorf("invalid assemble type(s): %s", invalidAsmTypes)
	}

	return o, nil
}

type Worker struct {
	ID uuid.UUID

	state stat.WorkerState

	listener   *Listener
	assemblers []assembler.Assembler

	packetStorage   *storage.PacketStorage
	assembleStorage *storage.AssembleStorage

	cancel func()
	lock   sync.Mutex
}

func NewWorker(ctx context.Context, opts *WorkerOptions) (w *Worker, c context.Context, err error) {
	opts, err = opts.Validate()
	if err != nil {
		return nil, nil, err
	}

	id := uuid.New()

	path, err := makeNamespace(id)
	if err != nil {
		return nil, nil, errors.Wrap(err, "worker: creating namespace")
	}

	assembleStorage := storage.NewAssembleStorage(path)
	packetStorage, err := storage.NewPacketStorage(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "worker: creating packet storage")
	}

	for _, t := range opts.CaptureLayers {
		s := storagefactory.New(t)
		if err := packetStorage.Register(t, s); err != nil {
			return nil, nil, errors.Wrap(err, "worker: registering packet layer")
		}
	}

	assemblers := make([]assembler.Assembler, len(opts.AssembleTypes))
	for idx, t := range opts.AssembleTypes {
		assemblers[idx] = asmfactory.New(t)

		if err := assembleStorage.Register(t); err != nil {
			return nil, nil, errors.Wrap(err, "worker: registering asm to storage")
		}

	}

	listener, err := newListener(&opts.ListenOptions)
	if err != nil {
		return nil, nil, err
	}

	c, cancel := context.WithCancel(ctx)

	w = &Worker{
		ID:              id,
		state:           stat.WorkerStateUp,
		listener:        listener,
		assemblers:      assemblers,
		assembleStorage: assembleStorage,
		packetStorage:   packetStorage,
		cancel:          cancel,
	}

	return
}

func (w *Worker) Exec(ctx context.Context) error {
	defer w.updateState(stat.WorkerStateFin)

	packets, err := w.listener.Listen(ctx)
	if err != nil {
		return err
	}

	var packet container.Packet
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case p, ok := <-packets:
			if !ok {
				log.Println("done")
				return nil
			}
			packet = p
		}

		if err := w.packetStorage.Store(ctx, packet); err != nil {
			w.Cancel()
			return errors.Wrap(err, "worker: storing the packet")
		}

		if tcpPacket, ok := packet.Layer.(*layers.TCP); ok {
			for _, asm := range w.assemblers {
				asm.Provide(tcpPacket)
			}
		}
	}
}

func (w *Worker) Cancel() {
	if w.updateState(stat.WorkerStateCancel) {
		w.cancel()
	}
}

func (w *Worker) updateState(s stat.WorkerState) (changed bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.state == stat.WorkerStateUp {
		w.state = s
		return true
	}
	return false
}

func makeNamespace(id uuid.UUID) (string, error) {
	path := filepath.Join(config.DefaultDataPath, id.String())
	return path, os.MkdirAll(path, 0644)
}
