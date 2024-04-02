package http

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/uuid"
	"github.com/onee-only/netrat/internal/container"
	"github.com/onee-only/netrat/internal/storage"
)

type httpStreamFactory struct {
	pctx context.Context

	connPairer *connPairer
	asmStorage storage.AssembleObjectStorage
}

// Only supports HTTP/1.X
// TODO: Support HTTPS
type httpStream struct {
	id       uuid.UUID
	isServer bool

	net, transport gopacket.Flow

	reasmStream chan tcpassembly.Reassembly
	buffered    tcpassembly.Reassembly

	isReading bool
	readLock  sync.Mutex

	firstSeen, lastSeen time.Time

	asmStorage storage.AssembleObjectStorage

	ctx    context.Context
	cancel func()
}

var _ tcpassembly.Stream = (*httpStream)(nil)

func (factory *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	sid, found := factory.connPairer.pairOrNew([2]gopacket.Flow{net, transport})

	s := &httpStream{
		id:  sid,
		net: net, transport: transport,
		isServer: found,

		reasmStream: make(chan tcpassembly.Reassembly),
		asmStorage:  factory.asmStorage,
	}

	s.ctx, s.cancel = context.WithCancel(factory.pctx)

	go s.readHTTP()

	return s
}

func (s *httpStream) readHTTP() {
	r := bufio.NewReader(s)

	resetReadable := func() {
		s.readLock.Lock()
		s.isReading = false
		s.readLock.Unlock()
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		var w interface{ Write(w io.Writer) error }
		if s.isServer {
			res, err := http.ReadResponse(r, nil)
			if err != nil {
				resetReadable()
				continue
			}
			w = res
		} else {
			req, err := http.ReadRequest(r)
			if err != nil {
				resetReadable()
				continue
			}
			w = req
		}

		s.lastSeen = s.buffered.Seen
		w.Write(os.Stdout)

		b := new(bytes.Buffer)
		w.Write(b)

		go func() {
			err := s.asmStorage.Store(context.WithoutCancel(s.ctx), container.Assembly{
				Object: b,
				Metadata: container.HTTPAsmMetadata{
					ID:         uuid.New(),
					StreamID:   s.id,
					Net:        s.net,
					Transport:  s.transport,
					Start:      s.firstSeen,
					End:        s.lastSeen,
					IsResponse: s.isServer,
				},
			})
			if err != nil {
				panic(err)
			}
		}()
		resetReadable()
	}
}

func (s *httpStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	for _, r := range reassemblies {
		if r.Start {
			// we ignore it when SYN flag is set.
			continue
		}

		s.readLock.Lock()
		if !s.isReading {
			s.isReading = true
			s.firstSeen = r.Seen
		}
		s.readLock.Unlock()

		select {
		case <-s.ctx.Done():
			return
		case s.reasmStream <- r:
		}

	}
}

func (s *httpStream) ReassemblyComplete() {
	s.cleanUp()
}

func (s *httpStream) Read(p []byte) (int, error) {
	if len(s.buffered.Bytes) == 0 {
		var ok bool
		if s.buffered, ok = <-s.reasmStream; !ok {
			return 0, io.EOF
		}
	}

	length := copy(p, s.buffered.Bytes)

	s.buffered.Bytes = s.buffered.Bytes[length:]
	return length, nil
}

func (s *httpStream) cleanUp() {
	s.cancel()
}
