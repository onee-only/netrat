package container

import (
	"github.com/google/gopacket"
	"github.com/google/uuid"
)

type Packet struct {
	gopacket.Packet

	ID uuid.UUID
}
