package assembly

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onee-only/netrat/internal/container"
	"github.com/onee-only/netrat/internal/storage"
	"github.com/onee-only/netrat/pkg/assemble"
	"github.com/onee-only/netrat/pkg/util"
	"github.com/pkg/errors"
)

type HTTPAsmStorage struct {
	path string
	db   *sqlx.DB
}

const httpTable = `
CREATE TABLE http(
	id BLOB PRIMARY KEY NOT NULL,
	sid BLOB NOT NULL,
	src BLOB NOT NULL, dst BLOB NOT NULL,
	start DATETIME NOT NULL, end DATETIME NOT NULL,
	is_response INT2 NOT NULL
)`

var _ storage.AssembleObjectStorage = (*HTTPAsmStorage)(nil)

func (s *HTTPAsmStorage) Init(db *sqlx.DB, base string) error {
	s.db = db

	if _, err := s.db.Exec(httpTable); err != nil {
		return errors.Wrap(err, "http assembly storage: creating http table")
	}

	s.path = filepath.Join(base, string(assemble.AssembleTypeHTTP))

	if err := os.Mkdir(s.path, 0644); err != nil {
		return errors.Wrap(err, "http assembly storage: creating dir")
	}

	return nil
}

func (s *HTTPAsmStorage) Store(ctx context.Context, asm container.Assembly) error {
	b := asm.Object
	metadata := asm.Metadata.(container.HTTPAsmMetadata)

	schema := httpToSchema(metadata)

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO http VALUES(
			:id, :sid, :src, :dst, 
			:start, :end, :is_response
		)`, schema)
	if err != nil {
		return errors.Wrap(err, "http assembly storage: storing metadata")
	}

	path := filepath.Join(s.path, metadata.ID.String())
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "http assembly storage: creating file")
	}
	defer f.Close()

	if _, err := f.ReadFrom(b); err != nil {
		return errors.Wrap(err, "http assembly storage: writing to file")
	}

	return nil
}

type HTTPSchema struct {
	ID         []byte    `db:"id"`
	SID        []byte    `db:"sid"`
	Src        string    `db:"src"`
	Dst        string    `db:"dst"`
	Start      time.Time `db:"start"`
	End        time.Time `db:"end"`
	IsResponse uint8     `db:"is_response"`
}

func httpToSchema(metadata container.HTTPAsmMetadata) (schema *HTTPSchema) {
	return &HTTPSchema{
		ID:         metadata.ID[:],
		SID:        metadata.StreamID[:],
		Src:        util.EndpointToString(metadata.Net.Src(), metadata.Transport.Src()),
		Dst:        util.EndpointToString(metadata.Net.Dst(), metadata.Transport.Dst()),
		Start:      metadata.Start,
		End:        metadata.End,
		IsResponse: util.BoolToUint8(metadata.IsResponse),
	}
}
