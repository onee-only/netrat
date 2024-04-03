package assembler

import (
	"github.com/onee-only/netrat/internal/container"
)

type Assembler interface {
	Provide(packet container.Packet)
	Valid(packet container.Packet) bool
}
