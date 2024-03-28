package container

import (
	"time"

	"github.com/google/gopacket"
	"github.com/google/uuid"
)

type Packet struct {
	ID        uuid.UUID
	Layer     gopacket.Layer
	Timestamp time.Time
}
