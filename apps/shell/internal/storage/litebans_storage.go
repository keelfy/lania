package storage

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
)

// LiteBans punishment kinds, named after their tables.
const (
	PunishmentKindBan  = "bans"
	PunishmentKindMute = "mutes"
)

// punishmentsPerKind caps the rows read from each table.
const punishmentsPerKind = 50

// PunishmentRecord is a row of litebans_bans or litebans_mutes.
type PunishmentRecord struct {
	ID int64
	// Kind is one of the PunishmentKind constants.
	Kind   string
	MCUUID uuid.UUID
	Reason string
	ByUUID string
	ByName string
	TimeMs int64
	// UntilMs is zero or less for a permanent punishment.
	UntilMs int64
	// Active is false once the punishment is lifted. LiteBans may keep it true after a temporary punishment runs out.
	Active        bool
	RemovedByUUID string
	RemovedByName string
	RemovedReason string
	RemovedAtMs   *int64
}

// LiteBansStorage reads the punishments of LiteBans. LiteBans is the only writer.
type LiteBansStorage interface {
	// FindPunishments returns the bans and mutes of the players, newest first, at most punishmentsPerKind of each
	// kind. Bans by IP alone have no uuid and are left out.
	FindPunishments(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*PunishmentRecord, error)
}

type liteBansStorage struct {
	db *stdsql.DB
}

func NewLiteBansStorage(db *stdsql.DB) LiteBansStorage {
	return &liteBansStorage{db: db}
}

// removed_by_date is read as epoch seconds: a TIMESTAMP scanned with parseTime would take the session time zone
// for UTC.
const findPunishmentsOfKind = `
(SELECT '%[1]s', id, uuid, COALESCE(reason, ''), COALESCE(banned_by_uuid, ''), COALESCE(banned_by_name, ''),
	time, until, active <> 0, COALESCE(removed_by_uuid, ''), COALESCE(removed_by_name, ''),
	COALESCE(removed_by_reason, ''), CAST(UNIX_TIMESTAMP(removed_by_date) * 1000 AS SIGNED)
FROM %[2]s
WHERE uuid IN (%[3]s)
ORDER BY time DESC
LIMIT %[4]d)`

func (s *liteBansStorage) FindPunishments(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*PunishmentRecord, error) {
	if len(mcUUIDs) == 0 {
		return nil, nil
	}

	placeholders, uuids := uuidArgs(mcUUIDs)
	query := fmt.Sprintf(findPunishmentsOfKind, PunishmentKindBan, liteBansTable(PunishmentKindBan), placeholders, punishmentsPerKind) +
		"\nUNION ALL" +
		fmt.Sprintf(findPunishmentsOfKind, PunishmentKindMute, liteBansTable(PunishmentKindMute), placeholders, punishmentsPerKind) +
		"\nORDER BY time DESC"
	args := append(append([]any{}, uuids...), uuids...)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*PunishmentRecord
	for rows.Next() {
		record := &PunishmentRecord{}
		var removedAt stdsql.NullInt64
		if err := rows.Scan(
			&record.Kind, &record.ID, &record.MCUUID, &record.Reason, &record.ByUUID, &record.ByName,
			&record.TimeMs, &record.UntilMs, &record.Active, &record.RemovedByUUID, &record.RemovedByName,
			&record.RemovedReason, &removedAt,
		); err != nil {
			return nil, err
		}
		if removedAt.Valid {
			record.RemovedAtMs = &removedAt.Int64
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func liteBansTable(name string) string {
	return fmt.Sprintf("`%s%s`", config.GetLiteBansTablePrefix(), name)
}
