package storage

import (
	"context"
	"database/sql"
	stdsql "database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/logger"
)

func newMySQLStorage(ctx context.Context, databaseName string, required bool) (*stdsql.DB, func(), error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.GetDatabaseUser(), config.GetDatabasePassword(), config.GetDatabaseHost(), config.GetDatabasePort(), databaseName)
	if !strings.Contains(dsn, "parseTime=") {
		if strings.Contains(dsn, "?") {
			dsn += "&parseTime=true"
		} else {
			dsn += "?parseTime=true"
		}
	}

	cleanup := func() {}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, cleanup, err
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		if required {
			return nil, cleanup, err
		}
		logger.Errorf(ctx, "failed to ping connection to %s: %v", databaseName, err)
	} else {
		logger.Infof(ctx, "Connection to %s (MySQL) established", databaseName)

		cleanup = func() {
			_ = db.Close()
		}
	}

	return db, cleanup, nil
}

func beginTx[T any](ctx context.Context, db *stdsql.DB, createQueries func(*stdsql.Tx) T, fn func(T) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Errorf(ctx, "failed to begin transaction: %v", err)
		return err
	}

	q := createQueries(tx)

	if err := fn(q); err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			logger.Errorf(ctx, "failed to rollback transaction: %v", err1)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Errorf(ctx, "failed to commit transaction: %v", err)
		return err
	}
	return nil
}

func ping(ctx context.Context, db *stdsql.DB) error {
	if err := db.PingContext(ctx); err != nil {
		logger.Debugf(ctx, "[SQL] Error pinging MySQL: %v", err)
		return err
	}
	return nil
}
