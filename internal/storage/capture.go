package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type CaptureStorage struct {
	path string
	db   *sqlx.DB
}

func NewCaptureStorage(path string) (*CaptureStorage, error) {
	storage := &CaptureStorage{
		path: path,
	}

	db, err := sqlx.Open("sqlite3", fmt.Sprintf("file:%s/captured.db", path))
	if err != nil {
		return nil, errors.Wrap(err, "capture storage: opening db")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "capture storage: ping db")
	}

	storage.db = db

	return storage, nil
}
