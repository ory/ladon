// +build !go1.8

package sql

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func connectContext(ctx context.Context, driver, connection string) (*sqlx.DB, error) {
	type result struct {
		db  *sqlx.DB
		err error
	}
	rc := make(chan *result, 1)
	go func() {
		db, err := sqlx.Connect(driver, connection)
		rc <- result{db: db, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-rc:
		return res.db, res.err
	}
}
