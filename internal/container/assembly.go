package container

import (
	"io"
	"time"

	"github.com/google/gopacket"
	"github.com/google/uuid"
)

type Assembly struct {
	Object   io.Reader
	Metadata any
}

type HTTPAsmMetadata struct {
	ID             uuid.UUID
	StreamID       uuid.UUID
	Net, Transport gopacket.Flow
	Start, End     time.Time
	IsResponse     bool
}
