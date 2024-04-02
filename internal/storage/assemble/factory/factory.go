package factory

import (
	"github.com/onee-only/netrat/internal/storage"
	"github.com/onee-only/netrat/internal/storage/assemble/assembly"
	"github.com/onee-only/netrat/pkg/assemble"
)

func New(t assemble.AssembleType) storage.AssembleObjectStorage {
	switch t {
	case assemble.AssembleTypeHTTP:
		return &assembly.HTTPAsmStorage{}
	}

	return nil
}
