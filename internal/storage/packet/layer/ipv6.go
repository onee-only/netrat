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

const ipv6Table = `
CREATE TABLE ipv6(
	id BLOB PRIMARY KEY NOT NULL,
	version INT NOT NULL, 
	priority INT NOT NULL,
	flow_label INT NOT NULL, 
	length INT NOT NULL,
	next_header INT NOT NULL, 
	hop_limit INT NOT NULL,
	src BLOB NOT NULL, dst BLOB NOT NULL
)`

type IPv6Storage struct{ db *sqlx.DB }

var _ storage.LayerStorage = (*IPv6Storage)(nil)

func (s *IPv6Storage) Init(db *sqlx.DB) error {
	_, err := db.Exec(ipv6Table)
	if err != nil {
		return errors.Wrap(err, "ipv6 storage: creating ipv6 table")
	}
	s.db = db
	return nil
}

func (s *IPv6Storage) Store(ctx context.Context, packet container.Packet) error {
	ipv6 := packet.NetworkLayer().(*layers.IPv6)
	schema := ipv6ToSchema(packet.ID, ipv6)

	_, err := s.db.NamedExecContext(ctx,
		`INSERT INTO ipv6 VALUES(
			:id, :version, :priority, :flow_label,
			:length, :next_header, :hop_limit,
			:src, :dst
		)`, schema)
	if err != nil {
		return errors.Wrap(err, "ipv6 storage: inserting ipv6 packet")
	}

	return nil
}

type IPv6Schema struct {
	ID         []byte `db:"id"`
	Version    uint8  `db:"version"`
	Priority   uint8  `db:"priority"`
	FlowLabel  uint32 `db:"flow_label"`
	Length     uint16 `db:"length"`
	NextHeader uint8  `db:"next_header"`
	HopLimit   uint8  `db:"hop_limit"`

	Src []byte `db:"src"`
	Dst []byte `db:"dst"`
}

func ipv6ToSchema(id uuid.UUID, ipv6 *layers.IPv6) (schema *IPv6Schema) {
	return &IPv6Schema{
		ID:         id[:],
		Version:    ipv6.Version,
		Priority:   ipv6.TrafficClass,
		FlowLabel:  ipv6.FlowLabel,
		Length:     ipv6.Length,
		NextHeader: uint8(ipv6.NextHeader),
		HopLimit:   ipv6.HopLimit,
		Src:        ipv6.SrcIP, Dst: ipv6.DstIP,
	}
}
