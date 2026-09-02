package fencedstore

import (
	"context"
	"database/sql"

	"github.com/Clint-Mathews/fencelock/lock"
)

// Postgres is a FencedResource backed by a Postgres table with a
// per-key last_token column, enforced with a conditional UPDATE.
type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{
		db: db,
	}
}

// Schema:
//
//	CREATE TABLE fenced_resource (
//		key 		TEXT PRIMARY KEY,
//		last_token	BIGINT NOT NULL DEFAULT 0,
//		data		BYTEA
//	);
func (p *Postgres) Write(ctx context.Context, key string, token int64, payload []byte) error {
	res, err := p.db.ExecContext(ctx, `
		INSERT INTO fenced_resource (key, last_token, data)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE
			SET last_token = EXCLUDED.last_token, data = EXCLUDED.data
			WHERE fenced_resource.last_token <= $2 
	`, key, token, payload)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return lock.ErrStaleToken
	}
	return nil
}
