package layer

import (
	"context"

	"github.com/google/gopacket/layers"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/onee-only/netrat/internal/container"
	"github.com/onee-only/netrat/internal/storage"
	"github.com/onee-only/netrat/pkg/util"
	"github.com/pkg/errors"
)

const tcpTable = `
CREATE TABLE tcp(
	id BLOB PRIMARY KEY NOT NULL,
	src INT NOT NULL, dst INT NOT NULL,
	seqnum INT NOT NULL, acknum INT NOT NULL,
	offset INT NOT NULL,

	fin INT2 NOT NULL, syn INT2 NOT NULL, rst INT2 NOT NULL, 
	psh INT2 NOT NULL, ack INT2 NOT NULL, urg INT2 NOT NULL,
	ece INT2 NOT NULL, cwr INT2 NOT NULL, ns INT2 NOT NULL,

	window INT NOT NULL, 
	checksum INT NOT NULL,
	urgent INT NOT NULL
)`

type TCPStorage struct{ db *sqlx.DB }

var _ storage.LayerStorage = (*TCPStorage)(nil)

func (s *TCPStorage) Init(db *sqlx.DB) error {
	_, err := db.Exec(tcpTable)
	if err != nil {
		return errors.Wrap(err, "tcp storage: creating tcp table")
	}
	s.db = db
	return nil
}

func (s *TCPStorage) Store(ctx context.Context, packet container.Packet) error {
	tcp := packet.TransportLayer().(*layers.TCP)
	schema := tcp2schema(packet.ID, tcp)
	_, err := s.db.NamedExecContext(ctx,
		`INSERT INTO tcp VALUES(
			:id, :src, :dst, :seqnum, :acknum, :offset,
			:fin, :syn, :rst, :psh, :ack, :urg, :ece, :cwr, :ns,
			:window, :checksum, :urgent
		)`, schema)
	if err != nil {
		return errors.Wrap(err, "tcp storage: inserting tcp packet")
	}

	return nil
}

type TcpSchema struct {
	ID     []byte `db:"id"`
	Src    uint16 `db:"src"`
	Dst    uint16 `db:"dst"`
	Seqnum uint32 `db:"seqnum"`
	Acknum uint32 `db:"acknum"`
	Offset uint8  `db:"offset"`

	Fin uint8 `db:"fin"`
	Syn uint8 `db:"syn"`
	Rst uint8 `db:"rst"`
	Psh uint8 `db:"psh"`
	Ack uint8 `db:"ack"`
	Urg uint8 `db:"urg"`
	Ece uint8 `db:"ece"`
	Cwr uint8 `db:"cwr"`
	Ns  uint8 `db:"ns"`

	Window   uint16 `db:"window"`
	Checksum uint16 `db:"checksum"`
	Urgent   uint16 `db:"urgent"`
}

func tcp2schema(id uuid.UUID, tcp *layers.TCP) (schema *TcpSchema) {
	return &TcpSchema{
		ID:  id[:],
		Src: uint16(tcp.SrcPort), Dst: uint16(tcp.DstPort),
		Seqnum: tcp.Seq, Acknum: tcp.Ack,
		Offset:   tcp.DataOffset,
		Fin:      util.BoolToUint8(tcp.FIN),
		Syn:      util.BoolToUint8(tcp.SYN),
		Rst:      util.BoolToUint8(tcp.RST),
		Psh:      util.BoolToUint8(tcp.PSH),
		Ack:      util.BoolToUint8(tcp.ACK),
		Urg:      util.BoolToUint8(tcp.URG),
		Ece:      util.BoolToUint8(tcp.ECE),
		Cwr:      util.BoolToUint8(tcp.CWR),
		Ns:       util.BoolToUint8(tcp.NS),
		Window:   tcp.Window,
		Checksum: tcp.Checksum,
		Urgent:   tcp.Urgent,
	}
}

func _(schema *TcpSchema) (tcp *layers.TCP) {
	return &layers.TCP{
		SrcPort: layers.TCPPort(schema.Src), DstPort: layers.TCPPort(schema.Dst),
		Seq: schema.Seqnum, Ack: schema.Acknum,
		DataOffset: schema.Offset,
		FIN:        util.Uint8ToBool(schema.Fin),
		SYN:        util.Uint8ToBool(schema.Syn),
		RST:        util.Uint8ToBool(schema.Rst),
		PSH:        util.Uint8ToBool(schema.Psh),
		ACK:        util.Uint8ToBool(schema.Ack),
		URG:        util.Uint8ToBool(schema.Urg),
		ECE:        util.Uint8ToBool(schema.Ece),
		CWR:        util.Uint8ToBool(schema.Cwr),
		NS:         util.Uint8ToBool(schema.Ns),
		Window:     schema.Window,
		Checksum:   schema.Checksum,
		Urgent:     schema.Urgent,
	}
}
