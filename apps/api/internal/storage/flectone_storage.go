package storage

import (
	"context"
	stdsql "database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lania-smp/backend/internal/config"
	storesql "github.com/lania-smp/backend/internal/storage/flectone"
)

type FlectoneStorage interface {
	Queries() storesql.Queries
	BeginTx(ctx context.Context, fn func(storesql.Queries) error) error
	Ping(ctx context.Context) error
}

type flectoneStorage struct {
	conn    *stdsql.DB
	queries storesql.Queries
}

func NewFlectoneStorage(ctx context.Context) (FlectoneStorage, func(), error) {
	db, cleanup, err := newMySQLStorage(ctx, config.GetDatabaseFlectoneName(), false)
	if err != nil {
		return nil, nil, err
	}

	sqlDatabase := &flectoneStorage{
		conn:    db,
		queries: storesql.NewMySQL(db),
	}
	return sqlDatabase, cleanup, nil
}

func (s *flectoneStorage) Queries() storesql.Queries {
	return s.queries
}

func (s *flectoneStorage) BeginTx(ctx context.Context, fn func(storesql.Queries) error) error {
	return beginTx(ctx, s.conn, storesql.WithMySQLTx, fn)
}

func (s *flectoneStorage) Ping(ctx context.Context) error {
	return ping(ctx, s.conn)
}
