package storagefactory

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/onee-only/netrat/internal/storage"
	"github.com/onee-only/netrat/internal/storage/packet/layer"
)

func New(t gopacket.LayerType) storage.LayerStorage {
	switch t {
	case layers.LayerTypeTCP:
		return &layer.TCPStorage{}
	case layers.LayerTypeIPv4:
		return &layer.IPv4Storage{}
	}

	return nil
}
