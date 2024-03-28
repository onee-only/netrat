package storage

import (
	"context"
	"fmt"

	"github.com/google/gopacket"
	"github.com/jmoiron/sqlx"
	"github.com/onee-only/netrat/internal/container"
	"github.com/pkg/errors"
)

type LayerStorage interface {
	Init(db *sqlx.DB) error
	Store(ctx context.Context, packet container.Packet) error
}

type PacketStorage struct {
	layerStorages map[gopacket.LayerType]LayerStorage

	db *sqlx.DB
}

func NewPacketStorage(path string) (*PacketStorage, error) {
	db, err := sqlx.Open("sqlite3", fmt.Sprintf("file:%s/packet.db", path))
	if err != nil {
		return nil, errors.Wrap(err, "packet storage: opening db")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "packet storage: ping db")
	}

	_, err = db.Exec(`
		CREATE TABLE packet(
			id BLOB NOT NULL PRIMARY KEY, 
			timestamp DATETIME NOT NULL,
			UNIQUE(id, timestamp)
		)`)
	if err != nil {
		return nil, errors.Wrap(err, "packet storage: creating packet table")
	}

	p := &PacketStorage{
		layerStorages: make(map[gopacket.LayerType]LayerStorage),

		db: db,
	}

	return p, nil
}

func (s *PacketStorage) Register(t gopacket.LayerType, storage LayerStorage) error {
	if err := storage.Init(s.db); err != nil {
		return errors.Wrap(err, "packet storage: layer storage init")
	}

	s.layerStorages[t] = storage

	return nil
}

func (s *PacketStorage) Store(ctx context.Context, packet container.Packet) error {
	ls := s.layerStorages[packet.Layer.LayerType()]

	_, err := s.db.ExecContext(ctx, "INSERT OR IGNORE INTO packet VALUES(?, ?)", packet.ID, packet.Timestamp)
	if err != nil {
		return errors.Wrap(err, "packet storage: inserting packet")
	}

	if err := ls.Store(ctx, packet); err != nil {
		return err
	}

	return nil
}

func (s *PacketStorage) Close() error {
	return s.db.Close()
}
