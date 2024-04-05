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

const (
	sectionTypeAnswer uint8 = iota
	sectionTypeAuthority
	sectionTypeAdditional
)

const dnsHeaderTable = `
CREATE TABLE dns_header(
	id BLOB PRIMARY KEY NOT NULL,
    tx_id INT NOT NULL,
    
	qr INT2 NOT NULL, op_code INT NOT NULL,
    aa INT2 NOT NULL, tc INT2 NOT NULL,
    rd INT2 NOT NULL, ra INT2 NOT NULL,
    
	z INT NOT NULL,
    res_code INT NOT NULL,
    qd_cnt INT NOT NULL, an_cnt INT NOT NULL,
    ns_cnt INT NOT NULL, ar_cnt INT NOT NULL,
    
	q_name BLOB NOT NULL,
    q_type INT NOT NULL,
    q_class INT NOT NULL
)`

const dnsRecordTable = `
CREATE TABLE dns_record(
	id BLOB NOT NULL,
	section INT NOT NULL, name BLOB NOT NULL,
	type INT NOT NULL, class INT NOT NULL,
	ttl INT NOT NULL, datalen INT NOT NULL,
	rdata BLOB NOT NULL
)`

type DNSStorage struct{ db *sqlx.DB }

var _ storage.LayerStorage = (*DNSStorage)(nil)

func (s *DNSStorage) Init(db *sqlx.DB) error {
	_, err := db.Exec(dnsHeaderTable)
	if err != nil {
		return errors.Wrap(err, "dns storage: creating dns_header table")
	}

	_, err = db.Exec(dnsRecordTable)
	if err != nil {
		return errors.Wrap(err, "dns storage: creating dns_record table")
	}

	s.db = db
	return nil
}

func (s *DNSStorage) Store(ctx context.Context, packet container.Packet) error {
	dns := packet.ApplicationLayer().(*layers.DNS)
	headerSchema := dnsHeaderToSchema(packet.ID, dns)

	if len(dns.Questions) == 0 {
		dns.Questions = append(dns.Questions, layers.DNSQuestion{})
	}

	_, err := s.db.NamedExecContext(ctx,
		`INSERT INTO dns_header VALUES(
			:id, :tx_id, :qr, :op_code, 
			:aa, :tc, :rd, :ra, :z, :res_code, 
			:qd_cnt, :an_cnt, :ns_cnt, :ar_cnt, 
			:q_name, :q_type, :q_class
		)`, headerSchema)
	if err != nil {
		return errors.Wrap(err, "dns storage: inserting dns packet")
	}

	for _, record := range dns.Answers {
		schema := dnsRecordToSchema(packet.ID, sectionTypeAnswer, &record)
		if err := insertDNSRecord(ctx, s.db, schema); err != nil {
			return err
		}
	}
	for _, record := range dns.Authorities {
		schema := dnsRecordToSchema(packet.ID, sectionTypeAuthority, &record)
		if err := insertDNSRecord(ctx, s.db, schema); err != nil {
			return err
		}
	}
	for _, record := range dns.Additionals {
		schema := dnsRecordToSchema(packet.ID, sectionTypeAdditional, &record)
		if err := insertDNSRecord(ctx, s.db, schema); err != nil {
			return err
		}
	}

	return nil
}

type DNSHeaderSchema struct {
	ID      []byte `db:"id"`
	TxID    uint16 `db:"tx_id"`
	QR      uint8  `db:"qr"`
	OpCode  uint8  `db:"op_code"`
	AA      uint8  `db:"aa"`
	TC      uint8  `db:"tc"`
	RD      uint8  `db:"rd"`
	RA      uint8  `db:"ra"`
	Z       uint8  `db:"z"`
	ResCode uint8  `db:"res_code"`
	QDCnt   uint16 `db:"qd_cnt"`
	ANCnt   uint16 `db:"an_cnt"`
	NSCnt   uint16 `db:"ns_cnt"`
	ARCnt   uint16 `db:"ar_cnt"`

	QName  []byte `db:"q_name"`
	QType  uint16 `db:"q_type"`
	QClass uint16 `db:"q_class"`
}

type DNSRecordSchema struct {
	ID      []byte `db:"id"`
	Section uint8  `db:"section"`
	Name    []byte `db:"name"`
	Type    uint16 `db:"type"`
	Class   uint16 `db:"class"`
	TTL     uint32 `db:"ttl"`
	DataLen uint16 `db:"datalen"`
	RData   []byte `db:"rdata"`
}

func insertDNSRecord(ctx context.Context, db *sqlx.DB, schema *DNSRecordSchema) error {
	_, err := db.NamedExecContext(ctx,
		`INSERT INTO dns_record VALUES(
			:id, :section,:name, :type, 
			:class, :ttl, :datalen, :rdata
		)`, schema)
	if err != nil {
		return errors.Wrap(err, "dns storage: inserting dns record")
	}
	return nil
}

func dnsHeaderToSchema(id uuid.UUID, dns *layers.DNS) (schema *DNSHeaderSchema) {
	return &DNSHeaderSchema{
		ID:      id[:],
		TxID:    dns.ID,
		QR:      util.BoolToUint8(dns.QR),
		OpCode:  uint8(dns.OpCode),
		AA:      util.BoolToUint8(dns.AA),
		TC:      util.BoolToUint8(dns.TC),
		RD:      util.BoolToUint8(dns.RD),
		RA:      util.BoolToUint8(dns.RA),
		Z:       dns.Z,
		ResCode: uint8(dns.ResponseCode),
		QDCnt:   dns.QDCount,
		ANCnt:   dns.ANCount,
		NSCnt:   dns.NSCount,
		ARCnt:   dns.ARCount,
		QName:   dns.Questions[0].Name,
		QType:   uint16(dns.Questions[0].Type),
		QClass:  uint16(dns.Questions[0].Class),
	}
}

func dnsRecordToSchema(id uuid.UUID, section uint8, rr *layers.DNSResourceRecord) (schema *DNSRecordSchema) {
	return &DNSRecordSchema{
		ID:      id[:],
		Section: section,
		Name:    rr.Name,
		Type:    uint16(rr.Type),
		Class:   uint16(rr.Class),
		TTL:     rr.TTL,
		DataLen: rr.DataLength,
		RData:   rr.Data,
	}
}
