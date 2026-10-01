package rpc

import (
	"context"

	"github.com/lania-smp/shell/internal/domain"
	shellv1 "github.com/lania-smp/shell/internal/gen/lania/shell/v1"
	"github.com/lania-smp/shell/internal/services"
)

type PunishmentHandler struct {
	shellv1.UnimplementedPunishmentServiceServer
	punishmentService services.PunishmentService
}

func NewPunishmentHandler(punishmentService services.PunishmentService) *PunishmentHandler {
	return &PunishmentHandler{punishmentService: punishmentService}
}

func (h *PunishmentHandler) GetPlayerPunishments(ctx context.Context, req *shellv1.GetPlayerPunishmentsRequest) (*shellv1.GetPlayerPunishmentsResponse, error) {
	mcUUIDs, err := parseUUIDs(req.GetMinecraftUuids())
	if err != nil {
		return nil, err
	}

	punishments, err := h.punishmentService.GetPlayerPunishments(ctx, mcUUIDs)
	if err != nil {
		return nil, internalError(err)
	}

	res := &shellv1.GetPlayerPunishmentsResponse{Punishments: make([]*shellv1.Punishment, len(punishments))}
	for i, punishment := range punishments {
		res.Punishments[i] = toPunishment(punishment)
	}
	return res, nil
}

var punishmentKinds = map[domain.PunishmentKind]shellv1.PunishmentKind{
	domain.PunishmentKindBan:  shellv1.PunishmentKind_PUNISHMENT_KIND_BAN,
	domain.PunishmentKindMute: shellv1.PunishmentKind_PUNISHMENT_KIND_MUTE,
}

var punishmentStatuses = map[domain.PunishmentStatus]shellv1.PunishmentStatus{
	domain.PunishmentStatusActive:  shellv1.PunishmentStatus_PUNISHMENT_STATUS_ACTIVE,
	domain.PunishmentStatusExpired: shellv1.PunishmentStatus_PUNISHMENT_STATUS_EXPIRED,
	domain.PunishmentStatusRemoved: shellv1.PunishmentStatus_PUNISHMENT_STATUS_REMOVED,
}

func toPunishment(punishment *domain.Punishment) *shellv1.Punishment {
	return &shellv1.Punishment{
		Id:            punishment.ID,
		Kind:          punishmentKinds[punishment.Kind],
		MinecraftUuid: punishment.MCUUID.String(),
		Reason:        punishment.Reason,
		IssuedBy:      punishment.IssuedBy,
		IssuedAtMs:    punishment.IssuedAtMs,
		ExpiresAtMs:   punishment.ExpiresAtMs,
		Status:        punishmentStatuses[punishment.Status],
		RemovedBy:     punishment.RemovedBy,
		RemovedReason: punishment.RemovedReason,
		RemovedAtMs:   punishment.RemovedAtMs,
	}
}
