package storage

import (
	"context"
	stdsql "database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lania-smp/backend/internal/config"
	storesql "github.com/lania-smp/backend/internal/storage/plan"
)

type PlanStorage interface {
	Queries() storesql.Queries
	BeginTx(ctx context.Context, fn func(storesql.Queries) error) error
	Ping(ctx context.Context) error
}

type planStorage struct {
	conn    *stdsql.DB
	queries storesql.Queries
}

func NewPlanStorage(ctx context.Context) (PlanStorage, func(), error) {
	db, cleanup, err := newMySQLStorage(ctx, config.GetDatabasePlanName(), false)
	if err != nil {
		return nil, nil, err
	}

	sqlDatabase := &planStorage{
		conn:    db,
		queries: storesql.NewMySQL(db),
	}
	return sqlDatabase, cleanup, nil
}

func (s *planStorage) Queries() storesql.Queries {
	return s.queries
}

func (s *planStorage) BeginTx(ctx context.Context, fn func(storesql.Queries) error) error {
	return beginTx(ctx, s.conn, storesql.WithMySQLTx, fn)
}

func (s *planStorage) Ping(ctx context.Context) error {
	return ping(ctx, s.conn)
}
