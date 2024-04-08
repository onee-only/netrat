package assembler

import (
	"github.com/onee-only/netrat/internal/container"
	"github.com/onee-only/netrat/pkg/assemble"
)

type Assembler interface {
	Provide(packet container.Packet)
	Valid(packet container.Packet) bool
	Type() assemble.AssembleType
}
