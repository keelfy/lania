package storage

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
)

// NavAuthStorage reads and writes the accounts of the NavAuth login plugin on the proxy. NavAuth keeps them in its
// own database on the same server. It reads credentials on every login and keeps no copy, so a change here
// applies on the next login.
type NavAuthStorage interface {
	// FindCredentialsRequired tells whether the player logs in with a password. found is false for a player
	// NavAuth never saw.
	FindCredentialsRequired(ctx context.Context, mcUUID uuid.UUID) (required bool, found bool, err error)
	// SavePassword stores the bcrypt hash as the password of the player and removes its two-factor secret.
	SavePassword(ctx context.Context, mcUUID uuid.UUID, passwordBcrypt string) error
}

type navAuthStorage struct {
	db *stdsql.DB
}

func NewNavAuthStorage(db *stdsql.DB) NavAuthStorage {
	return &navAuthStorage{db: db}
}

const findNavAuthCredentialsRequired = `
SELECT credentials_required
FROM %s
WHERE uuid = ?
`

func (s *navAuthStorage) FindCredentialsRequired(ctx context.Context, mcUUID uuid.UUID) (bool, bool, error) {
	var required bool
	query := fmt.Sprintf(findNavAuthCredentialsRequired, navAuthTable(config.GetNavAuthUsersTableName()))
	err := s.db.QueryRowContext(ctx, query, mcUUID.String()).Scan(&required)
	if errors.Is(err, stdsql.ErrNoRows) {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}
	return required, true, nil
}

// NavAuth names the columns after its fields and keeps the algorithm as the enum name.
const saveNavAuthPassword = `
INSERT INTO %s (uuid, passwordHash, algo, twoFactorSecret)
VALUES (?, ?, 'BCRYPT', NULL)
ON DUPLICATE KEY UPDATE passwordHash = VALUES(passwordHash), algo = VALUES(algo), twoFactorSecret = NULL
`

func (s *navAuthStorage) SavePassword(ctx context.Context, mcUUID uuid.UUID, passwordBcrypt string) error {
	query := fmt.Sprintf(saveNavAuthPassword, navAuthTable(config.GetNavAuthCredentialsTableName()))
	_, err := s.db.ExecContext(ctx, query, mcUUID.String(), passwordBcrypt)
	return err
}

// navAuthTable qualifies the table with the NavAuth database.
func navAuthTable(table string) string {
	return fmt.Sprintf("`%s`.`%s`", config.GetNavAuthDatabaseName(), table)
}
