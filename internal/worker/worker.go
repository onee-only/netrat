package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/uuid"
	"github.com/onee-only/netrat/internal/assembler"
	asmfactory "github.com/onee-only/netrat/internal/assembler/factory"
	"github.com/onee-only/netrat/internal/config"
	"github.com/onee-only/netrat/internal/container"
	"github.com/onee-only/netrat/internal/storage"
	astoragefactory "github.com/onee-only/netrat/internal/storage/assemble/factory"
	pstoragefactory "github.com/onee-only/netrat/internal/storage/packet/factory"
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

	slices.Sort(o.AssembleTypes)
	o.AssembleTypes = slices.Compact(o.AssembleTypes)

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
	id      uuid.UUID
	start   time.Time
	timeout time.Duration

	state stat.WorkerState

	listener   *listener
	assemblers []assembler.Assembler

	packetStorage   *storage.PacketStorage
	assembleStorage *storage.AssembleStorage

	cancel func()
	lock   sync.Mutex
}

func New(ctx context.Context, opts *WorkerOptions) (w *Worker, c context.Context, err error) {
	opts, err = opts.Validate()
	if err != nil {
		return nil, nil, err
	}

	id := uuid.New()

	path, err := makeNamespace(id)
	if err != nil {
		return nil, nil, errors.Wrap(err, "worker: creating namespace")
	}

	capStorage, err := storage.NewCaptureStorage(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "worker: creating capture storage")
	}

	assembleStorage, err := storage.NewAssembleStorage(capStorage)
	if err != nil {
		return nil, nil, errors.Wrap(err, "worker: creating assemble storage")
	}

	packetStorage, err := storage.NewPacketStorage(capStorage)
	if err != nil {
		return nil, nil, errors.Wrap(err, "worker: creating packet storage")
	}

	for _, t := range opts.CaptureLayers {
		s := pstoragefactory.New(t)
		if err := packetStorage.Register(t, s); err != nil {
			return nil, nil, errors.Wrap(err, "worker: registering packet layer")
		}
	}

	assemblers := make([]assembler.Assembler, len(opts.AssembleTypes))
	for idx, t := range opts.AssembleTypes {
		s := astoragefactory.New(t)

		if err := assembleStorage.Register(t, s); err != nil {
			return nil, nil, errors.Wrap(err, "worker: registering asm to storage")
		}

		assemblers[idx] = asmfactory.New(t, s)
	}

	listener, err := newListener(&opts.ListenOptions)
	if err != nil {
		return nil, nil, err
	}

	c, cancel := context.WithCancel(ctx)

	w = &Worker{
		id:              id,
		timeout:         opts.Timeout,
		state:           stat.WorkerStateInit,
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

	packets, err := w.listener.listen(ctx)
	if err != nil {
		return err
	}

	w.state = stat.WorkerStateUp
	w.start = time.Now()

	var ok bool
	var packet container.Packet
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case packet, ok = <-packets:
			if !ok {
				log.Println("done")
				return nil
			}
		}

		if err := w.packetStorage.Store(ctx, packet); err != nil {
			w.Cancel()
			return errors.Wrap(err, "worker: storing the packet")
		}

		for _, asm := range w.assemblers {
			if asm.Valid(packet) {
				asm.Provide(packet)
			}
		}
	}
}

func (w *Worker) Cancel() {
	if w.updateState(stat.WorkerStateCancel) {
		w.cancel()
	}
}

func (w *Worker) ID() uuid.UUID {
	return w.id
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

func (w *Worker) ExportStats() stat.Worker {
	stat := stat.Worker{
		ID:        w.id,
		CreatedAt: w.start,
		Timeout:   w.timeout,
		State:     w.state,

		SnapLen:     w.listener.opts.SnapLen,
		Promiscuous: w.listener.opts.Promiscuous,
		Captures:    w.listener.opts.CaptureLayers,
		BPFFilter:   w.listener.opts.BPFFilter,
	}

	if w.listener.opts.Device != "" {
		stat.Live = true
		stat.Src = w.listener.opts.Device
	} else {
		stat.Src = w.listener.opts.PcapFile
	}

	assembles := make([]assemble.AssembleType, 0, len(w.assemblers))
	for _, asm := range w.assemblers {
		assembles = append(assembles, asm.Type())
	}

	stat.Assembles = assembles

	return stat
}

func makeNamespace(id uuid.UUID) (string, error) {
	path := filepath.Join(config.DefaultDataPath, id.String())
	return path, os.MkdirAll(path, 0644)
}
