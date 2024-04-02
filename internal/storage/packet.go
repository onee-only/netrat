package storage

import (
	"context"
	"time"

	"github.com/google/gopacket"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/onee-only/netrat/internal/container"
	"github.com/pkg/errors"
)

type LayerStorage interface {
	Init(db *sqlx.DB) error
	Store(ctx context.Context, packet container.Packet) error
}

type PacketStorage struct {
	db *sqlx.DB

	layerStorages map[gopacket.LayerType]LayerStorage
}

func NewPacketStorage(capStorage *CaptureStorage) (*PacketStorage, error) {
	storage := &PacketStorage{
		db: capStorage.db,

		layerStorages: make(map[gopacket.LayerType]LayerStorage),
	}

	_, err := storage.db.Exec(`
		CREATE TABLE packet(
			id BLOB NOT NULL PRIMARY KEY, 
			timestamp DATETIME NOT NULL,
			UNIQUE(id, timestamp),
			FOREIGN KEY(id) REFERENCES packet(id)
		)`)
	if err != nil {
		return nil, errors.Wrap(err, "packet storage: creating packet table")
	}

	return storage, nil
}

func (s *PacketStorage) Register(t gopacket.LayerType, storage LayerStorage) error {
	if err := storage.Init(s.db); err != nil {
		return errors.Wrap(err, "packet storage: layer storage init")
	}

	s.layerStorages[t] = storage

	return nil
}

func (s *PacketStorage) Store(ctx context.Context, packet container.Packet) error {
	if err := s.storeMetadata(ctx, packet.ID, packet.Metadata().Timestamp); err != nil {
		return err
	}

	for t, storage := range s.layerStorages {
		if layer := packet.Layer(t); layer != nil {
			if err := storage.Store(ctx, packet); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *PacketStorage) storeMetadata(ctx context.Context, id uuid.UUID, timestamp time.Time) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO packet VALUES(?, ?)", id[:], timestamp)
	if err != nil {
		return errors.Wrap(err, "packet storage: inserting packet")
	}
	return nil
}

func (s *PacketStorage) Close() error {
	return s.db.Close()
}
