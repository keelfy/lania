package storage

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
	"github.com/lania-smp/shell/internal/logger"
)

// NewDatabase opens the connection pool shared by every storage.
func NewDatabase(ctx context.Context) (*stdsql.DB, func(), error) {
	databaseName := config.GetDatabaseName()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.GetDatabaseUser(), config.GetDatabasePassword(), config.GetDatabaseHost(), config.GetDatabasePort(), databaseName)

	db, err := stdsql.Open("mysql", dsn)
	if err != nil {
		return nil, func() {}, err
	}
	cleanup := func() {
		_ = db.Close()
	}

	// The database may appear after shell starts, so an unreachable
	// database is logged instead of failing startup.
	if err := db.PingContext(ctx); err != nil {
		logger.Errorf(ctx, "failed to ping connection to %s: %v", databaseName, err)
	} else {
		logger.Infof(ctx, "Connection to %s (MySQL) established", databaseName)
	}

	return db, cleanup, nil
}

func beginTx(ctx context.Context, db *stdsql.DB, fn func(*stdsql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			logger.Errorf(ctx, "failed to rollback transaction: %v", err1)
		}
		return err
	}

	return tx.Commit()
}

// uuidArgs returns "?, ?, ..." placeholders and matching args for an IN clause.
func uuidArgs(mcUUIDs uuid.UUIDs) (string, []any) {
	placeholders := make([]string, len(mcUUIDs))
	args := make([]any, len(mcUUIDs))
	for i, mcUUID := range mcUUIDs {
		placeholders[i] = "?"
		args[i] = mcUUID.String()
	}
	return strings.Join(placeholders, ", "), args
}
