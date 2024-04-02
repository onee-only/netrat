package asmfactory

import (
	"github.com/onee-only/netrat/internal/assembler"
	"github.com/onee-only/netrat/internal/assembler/http"
	"github.com/onee-only/netrat/internal/storage"
	"github.com/onee-only/netrat/pkg/assemble"
)

func New(t assemble.AssembleType, storage storage.AssembleObjectStorage) assembler.Assembler {
	switch t {
	case assemble.AssembleTypeHTTP:
		return http.NewHTTPAssembler(storage)
	}

	return nil
}
