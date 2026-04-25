package flectonesql

import (
	"context"
	stdsql "database/sql"

	"github.com/google/uuid"
)

type Queries interface {
	// Profile Playtime
	FindOnlineByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
}

type queryable interface {
	ExecContext(ctx context.Context, query string, args ...any) (stdsql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*stdsql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *stdsql.Row
}

type queries struct {
	x queryable
}

// NewMySQL returns Queries backed by database/sql MySQL driver
func NewMySQL(db *stdsql.DB) Queries {
	return &queries{x: db}
}

// WithMySQLTx returns Queries bound to a single transaction
func WithMySQLTx(tx *stdsql.Tx) Queries {
	return &queries{x: tx}
}
