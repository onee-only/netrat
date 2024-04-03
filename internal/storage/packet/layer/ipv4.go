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

const ipv4Table = `
CREATE TABLE ipv4(
	id BLOB PRIMARY KEY NOT NULL,
	version INT NOT NULL, 
	hlen INT NOT NULL,
	tos INT NOT NULL, 
	length INT NOT NULL,

	frag_id INT NOT NULL, 
	frag_flag INT NOT NULL,
	frag_offset INT NOT NULL,

	ttl INT NOT NULL, 
	protocol INT NOT NULL,
	checksum INT NOT NULL,
	src BLOB NOT NULL, dst BLOB NOT NULL
)`

type IPv4Storage struct{ db *sqlx.DB }

var _ storage.LayerStorage = (*IPv4Storage)(nil)

func (s *IPv4Storage) Init(db *sqlx.DB) error {
	_, err := db.Exec(ipv4Table)
	if err != nil {
		return errors.Wrap(err, "ip storage: creating ip table")
	}
	s.db = db
	return nil
}

func (s *IPv4Storage) Store(ctx context.Context, packet container.Packet) error {
	ipv4 := packet.NetworkLayer().(*layers.IPv4)
	schema := ipv4ToSchema(packet.ID, ipv4)

	_, err := s.db.NamedExecContext(ctx,
		`INSERT INTO ipv4 VALUES(
			:id, :version, :hlen, :tos, :length,
			:frag_id, :frag_flag, :frag_offset,
			:ttl, :protocol, :checksum,
			:src, :dst
		)`, schema)
	if err != nil {
		return errors.Wrap(err, "ip storage: inserting ip packet")
	}

	return nil
}

type IPv4Schema struct {
	ID      []byte `db:"id"`
	Version uint8  `db:"version"`
	HLen    uint8  `db:"hlen"`
	TOS     uint8  `db:"tos"`
	Length  uint16 `db:"length"`

	FragID     uint16 `db:"frag_id"`
	FragFlag   uint8  `db:"frag_flag"`
	FragOffset uint16 `db:"frag_offset"`

	TTL      uint8  `db:"ttl"`
	Protocol uint8  `db:"protocol"`
	Checksum uint16 `db:"checksum"`

	Src []byte `db:"src"`
	Dst []byte `db:"dst"`
}

func ipv4ToSchema(id uuid.UUID, ipv4 *layers.IPv4) (schema *IPv4Schema) {
	return &IPv4Schema{
		ID:         id[:],
		Version:    ipv4.Version,
		HLen:       ipv4.IHL,
		TOS:        ipv4.TOS,
		Length:     ipv4.Length,
		FragID:     ipv4.Id,
		FragFlag:   uint8(ipv4.Flags),
		FragOffset: ipv4.FragOffset,
		TTL:        ipv4.TTL,
		Protocol:   uint8(ipv4.Protocol),
		Checksum:   ipv4.Checksum,
		Src:        ipv4.SrcIP, Dst: ipv4.DstIP,
	}
}
