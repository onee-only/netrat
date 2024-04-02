package storage

import (
	"context"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/onee-only/netrat/internal/container"
	"github.com/onee-only/netrat/pkg/assemble"
	"github.com/pkg/errors"
)

type AssembleObjectStorage interface {
	Init(db *sqlx.DB, base string) error
	Store(ctx context.Context, asm container.Assembly) error
}

type AssembleStorage struct {
	objectStorages map[assemble.AssembleType]AssembleObjectStorage

	db   *sqlx.DB
	base string
}

func NewAssembleStorage(capStorage *CaptureStorage) (*AssembleStorage, error) {
	storage := &AssembleStorage{
		objectStorages: make(map[assemble.AssembleType]AssembleObjectStorage),
		db:             capStorage.db,
	}

	storage.base = filepath.Join(capStorage.path, "asm")

	if err := os.Mkdir(storage.base, 0644); err != nil {
		return nil, errors.Wrap(err, "assemble storage: creating base dir")
	}

	return storage, nil
}

func (s *AssembleStorage) Register(t assemble.AssembleType, storage AssembleObjectStorage) error {
	if err := storage.Init(s.db, s.base); err != nil {
		return errors.Wrap(err, "assemble storage: registering object storage")
	}

	s.objectStorages[t] = storage

	return nil
}
