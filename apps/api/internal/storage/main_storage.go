package storage

import (
	"context"
	stdsql "database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lania-smp/backend/internal/config"
	storesql "github.com/lania-smp/backend/internal/storage/main"
)

type MainStorage interface {
	Queries() storesql.Queries
	BeginTx(ctx context.Context, fn func(storesql.Queries) error) error
	Ping(ctx context.Context) error
}

type mainStorage struct {
	conn    *stdsql.DB
	queries storesql.Queries
}

func NewMainStorage(ctx context.Context) (MainStorage, func(), error) {
	db, cleanup, err := newMySQLStorage(ctx, config.GetDatabaseName(), true)
	if err != nil {
		return nil, nil, err
	}

	sqlDatabase := &mainStorage{
		conn:    db,
		queries: storesql.NewMySQL(db),
	}
	return sqlDatabase, cleanup, nil
}

func (s *mainStorage) Queries() storesql.Queries {
	return s.queries
}

func (s *mainStorage) BeginTx(ctx context.Context, fn func(storesql.Queries) error) error {
	return beginTx(ctx, s.conn, storesql.WithMySQLTx, fn)
}

func (s *mainStorage) Ping(ctx context.Context) error {
	return ping(ctx, s.conn)
}
