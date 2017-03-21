// +build go1.8

package sql

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func connectContext(ctx context.Context, driver, connection string) (*sqlx.DB, error) {
	return sqlx.ConnectContext(ctx, driver, connection)
}
