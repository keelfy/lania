package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/domain"
	"github.com/lania-smp/shell/internal/storage"
)

type PunishmentService interface {
	// GetPlayerPunishments returns the bans and mutes of the players, newest first.
	GetPlayerPunishments(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*domain.Punishment, error)
	// GetBannedPlayers returns the players with a ban in force now, each once.
	GetBannedPlayers(ctx context.Context, mcUUIDs uuid.UUIDs) (uuid.UUIDs, error)
}

type punishmentService struct {
	liteBansStorage storage.LiteBansStorage
	now             func() time.Time
}

func NewPunishmentService(liteBansStorage storage.LiteBansStorage) PunishmentService {
	return &punishmentService{liteBansStorage: liteBansStorage, now: time.Now}
}

func (s *punishmentService) GetPlayerPunishments(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*domain.Punishment, error) {
	records, err := s.liteBansStorage.FindPunishments(ctx, mcUUIDs)
	if err != nil {
		return nil, err
	}

	nowMs := s.now().UnixMilli()
	punishments := make([]*domain.Punishment, 0, len(records))
	for _, record := range records {
		punishments = append(punishments, toPunishment(record, nowMs))
	}
	return punishments, nil
}

func (s *punishmentService) GetBannedPlayers(ctx context.Context, mcUUIDs uuid.UUIDs) (uuid.UUIDs, error) {
	return s.liteBansStorage.FindBannedPlayers(ctx, mcUUIDs, s.now().UnixMilli())
}

var punishmentKinds = map[string]domain.PunishmentKind{
	storage.PunishmentKindBan:  domain.PunishmentKindBan,
	storage.PunishmentKindMute: domain.PunishmentKindMute,
}

func toPunishment(record *storage.PunishmentRecord, nowMs int64) *domain.Punishment {
	punishment := &domain.Punishment{
		ID:            record.ID,
		Kind:          punishmentKinds[record.Kind],
		MCUUID:        record.MCUUID,
		Reason:        record.Reason,
		IssuedBy:      moderatorName(record.ByUUID, record.ByName),
		IssuedAtMs:    record.TimeMs,
		Status:        punishmentStatus(record, nowMs),
		RemovedReason: record.RemovedReason,
	}
	if record.UntilMs > 0 {
		punishment.ExpiresAtMs = &record.UntilMs
	}
	if punishment.Status == domain.PunishmentStatusRemoved {
		punishment.RemovedBy = moderatorName(record.RemovedByUUID, record.RemovedByName)
		punishment.RemovedAtMs = record.RemovedAtMs
	} else {
		punishment.RemovedReason = ""
	}
	return punishment
}

// punishmentStatus follows litebans-php: LiteBans may keep a temporary punishment active after it runs out, and
// marks the ones it closes itself with a remover starting with "#", like "#expired".
func punishmentStatus(record *storage.PunishmentRecord, nowMs int64) domain.PunishmentStatus {
	runOut := record.UntilMs > 0 && record.UntilMs <= nowMs
	switch {
	case isLiteBansMarker(record.RemovedByUUID) || isLiteBansMarker(record.RemovedByName):
		return domain.PunishmentStatusExpired
	case record.RemovedByUUID != "" || record.RemovedByName != "":
		return domain.PunishmentStatusRemoved
	case runOut:
		return domain.PunishmentStatusExpired
	case record.Active:
		return domain.PunishmentStatusActive
	default:
		// Lifted without a remover, for example replaced by a newer punishment.
		return domain.PunishmentStatusRemoved
	}
}

func isLiteBansMarker(value string) bool {
	return strings.HasPrefix(value, "#")
}

// moderatorName gives nil for the console: it has no player UUID.
func moderatorName(byUUID, byName string) *string {
	if byName == "" || strings.EqualFold(byName, "console") {
		return nil
	}
	if _, err := uuid.Parse(byUUID); err != nil {
		return nil
	}
	return &byName
}
