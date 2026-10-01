package services

import (
	"testing"

	"github.com/lania-smp/shell/internal/domain"
	"github.com/lania-smp/shell/internal/storage"
)

const moderatorUUID = "6a1d8f0e-0c43-4b47-9d57-8e2f4d2b3c11"

func TestPunishmentStatus(t *testing.T) {
	const now = int64(10_000)
	tests := []struct {
		name   string
		record storage.PunishmentRecord
		want   domain.PunishmentStatus
	}{
		{"permanent", storage.PunishmentRecord{Active: true}, domain.PunishmentStatusActive},
		{"temporary running", storage.PunishmentRecord{Active: true, UntilMs: now + 1}, domain.PunishmentStatusActive},
		{"temporary run out but still active", storage.PunishmentRecord{Active: true, UntilMs: now}, domain.PunishmentStatusExpired},
		{"closed by LiteBans", storage.PunishmentRecord{UntilMs: now - 1, RemovedByUUID: "#expired", RemovedByName: "#expired"}, domain.PunishmentStatusExpired},
		{"lifted by moderator", storage.PunishmentRecord{UntilMs: now + 1, RemovedByUUID: moderatorUUID, RemovedByName: "Steve"}, domain.PunishmentStatusRemoved},
		{"lifted by console", storage.PunishmentRecord{RemovedByUUID: "CONSOLE", RemovedByName: "Console"}, domain.PunishmentStatusRemoved},
		{"inactive permanent without remover", storage.PunishmentRecord{}, domain.PunishmentStatusRemoved},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := punishmentStatus(&tt.record, now); got != tt.want {
				t.Fatalf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestToPunishmentModerators(t *testing.T) {
	removed := toPunishment(&storage.PunishmentRecord{
		Kind: storage.PunishmentKindMute, ByUUID: "CONSOLE", ByName: "Console",
		RemovedByUUID: moderatorUUID, RemovedByName: "Steve", RemovedReason: "appeal", RemovedAtMs: ptr(int64(5)),
	}, 10)
	if removed.Kind != domain.PunishmentKindMute || removed.IssuedBy != nil || removed.ExpiresAtMs != nil {
		t.Fatalf("console mute = %+v", removed)
	}
	if removed.RemovedBy == nil || *removed.RemovedBy != "Steve" || removed.RemovedReason != "appeal" || removed.RemovedAtMs == nil {
		t.Fatalf("remover = %+v", removed)
	}

	expired := toPunishment(&storage.PunishmentRecord{
		Kind: storage.PunishmentKindBan, ByUUID: moderatorUUID, ByName: "Steve", UntilMs: 9,
		RemovedByUUID: "#expired", RemovedByName: "#expired", RemovedReason: "x",
	}, 10)
	if expired.IssuedBy == nil || *expired.IssuedBy != "Steve" || expired.ExpiresAtMs == nil {
		t.Fatalf("expired ban = %+v", expired)
	}
	if expired.RemovedBy != nil || expired.RemovedReason != "" || expired.RemovedAtMs != nil {
		t.Fatalf("expired ban has a remover: %+v", expired)
	}
}
