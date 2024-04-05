package layer

import (
	"context"

	"github.com/google/gopacket/layers"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/onee-only/netrat/internal/container"
	"github.com/onee-only/netrat/internal/storage"
	"github.com/pkg/errors"
)

const udpTable = `
CREATE TABLE udp(
	id BLOB PRIMARY KEY NOT NULL,
    src INT NOT NULL,
    dst INT NOT NULL,
    length INT NOT NULL,
    checksum INT NOT NULL
)`

type UDPStorage struct{ db *sqlx.DB }

var _ storage.LayerStorage = (*UDPStorage)(nil)

func (s *UDPStorage) Init(db *sqlx.DB) error {
	_, err := db.Exec(udpTable)
	if err != nil {
		return errors.Wrap(err, "udp storage: creating udp table")
	}
	s.db = db
	return nil
}

func (s *UDPStorage) Store(ctx context.Context, packet container.Packet) error {
	udp := packet.TransportLayer().(*layers.UDP)
	schema := udpToSchema(packet.ID, udp)

	_, err := s.db.NamedExecContext(ctx,
		`INSERT INTO udp VALUES(
			:id, :src, :dst, :length, :checksum
		)`, schema)
	if err != nil {
		return errors.Wrap(err, "udp storage: inserting udp packet")
	}

	return nil
}

type UDPSchema struct {
	ID       []byte `db:"id"`
	Src      uint16 `db:"src"`
	Dst      uint16 `db:"dst"`
	Length   uint16 `db:"length"`
	Checksum uint16 `db:"checksum"`
}

func udpToSchema(id uuid.UUID, udp *layers.UDP) (schema *UDPSchema) {
	return &UDPSchema{
		ID:       id[:],
		Src:      uint16(udp.SrcPort),
		Dst:      uint16(udp.DstPort),
		Length:   udp.Length,
		Checksum: udp.Checksum,
	}
}
