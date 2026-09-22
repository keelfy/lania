package services

import (
	"context"
	stdsql "database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

// MaxNotifications is the longest list the bell menu asks for.
const MaxNotifications = 50

type NotificationService interface {
	// GetNotifications returns the notifications of the user that match the filter, newest first,
	// and whether more of them follow past the end of the list.
	GetNotifications(ctx context.Context, userID uuid.UUID, filter domain.NotificationFilter) ([]*domain.Notification, bool, error)
	// CountUnreadNotifications returns how many notifications the user has not read yet.
	CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkNotificationsRead stamps the unread notifications of the user. Empty ids marks all of them.
	MarkNotificationsRead(ctx context.Context, userID uuid.UUID, ids uuid.UUIDs) error
	// NotifyCosmeticGranted tells the owner of the profile about a new name color or name prefix.
	// A profile without an owner is skipped. A failure is logged, not returned, so a notification
	// never undoes the grant it reports.
	NotifyCosmeticGranted(ctx context.Context, queries sql.Queries, profile *domain.Profile, grantType domain.GrantType, itemID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID)
	// NotifyCosmeticRevoked tells the owner of the profile that a name color or a name prefix was taken back.
	NotifyCosmeticRevoked(ctx context.Context, queries sql.Queries, profile *domain.Profile, grantType domain.GrantType, itemID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID)
	// NotifyProfileMerged tells the owner of the target profile that an admin moved sourceUsername's data into it.
	// A profile without an owner is skipped. A failure is logged, not returned, so a notification never undoes the merge.
	NotifyProfileMerged(ctx context.Context, queries sql.Queries, profile *domain.Profile, sourceUsername string)
}

type notificationService struct {
	storage storage.MainStorage
}

func NewNotificationService(storage storage.MainStorage) NotificationService {
	return &notificationService{storage: storage}
}

func (s *notificationService) GetNotifications(ctx context.Context, userID uuid.UUID, filter domain.NotificationFilter) ([]*domain.Notification, bool, error) {
	if filter.Limit <= 0 || filter.Limit > MaxNotifications {
		filter.Limit = MaxNotifications
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// One row past the limit tells whether there is more, without counting the whole table.
	limit := filter.Limit
	filter.Limit++
	notifications, err := s.storage.Queries().FindNotificationsByUserID(ctx, userID, filter)
	if err == stdsql.ErrNoRows {
		return []*domain.Notification{}, false, nil
	} else if err != nil {
		return nil, false, utils.NewInternalServerError("failed to find notifications", err)
	}
	if len(notifications) > limit {
		return notifications[:limit], true, nil
	}
	return notifications, false, nil
}

func (s *notificationService) CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.storage.Queries().CountUnreadNotificationsByUserID(ctx, userID)
	if err != nil {
		return 0, utils.NewInternalServerError("failed to count unread notifications", err)
	}
	return count, nil
}

func (s *notificationService) MarkNotificationsRead(ctx context.Context, userID uuid.UUID, ids uuid.UUIDs) error {
	if err := s.storage.Queries().MarkNotificationsRead(ctx, userID, ids); err != nil {
		return utils.NewInternalServerError("failed to mark notifications read", err)
	}
	return nil
}

func (s *notificationService) NotifyCosmeticGranted(ctx context.Context, queries sql.Queries, profile *domain.Profile, grantType domain.GrantType, itemID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) {
	s.notifyCosmetic(ctx, queries, domain.NotificationTypeCosmeticGranted, profile, grantType, itemID, prefixType, seasonID)
}

func (s *notificationService) NotifyCosmeticRevoked(ctx context.Context, queries sql.Queries, profile *domain.Profile, grantType domain.GrantType, itemID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) {
	s.notifyCosmetic(ctx, queries, domain.NotificationTypeCosmeticRevoked, profile, grantType, itemID, prefixType, seasonID)
}

func (s *notificationService) notifyCosmetic(
	ctx context.Context,
	queries sql.Queries,
	notificationType domain.NotificationType,
	profile *domain.Profile,
	grantType domain.GrantType,
	itemID uuid.UUID,
	prefixType domain.ProfilePrefixType,
	seasonID *uuid.UUID,
) {
	if profile == nil || profile.OwnerUserID == nil {
		return
	}

	payload := domain.CosmeticNotificationPayload{
		ProfileID:       profile.ID,
		ProfileUsername: profile.MinecraftUsername,
		GrantType:       grantType,
		PrefixType:      prefixType,
		ItemID:          itemID,
		SeasonID:        seasonID,
	}
	if err := s.fillCosmetic(ctx, queries, &payload); err != nil {
		logger.Errorf(ctx, "failed to read cosmetic %s for a notification: %v", itemID, err)
		return
	}
	if seasonID != nil {
		// The season name is a nicety: without it the notification still reads fine.
		if season, err := queries.FindSeasonByID(ctx, *seasonID); err != nil {
			logger.Errorf(ctx, "failed to read season %s for a notification: %v", *seasonID, err)
		} else {
			payload.SeasonName = season.Name
		}
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf(ctx, "failed to build the payload of a cosmetic notification: %v", err)
		return
	}

	err = queries.InsertNotification(ctx, sql.InsertNotificationParams{
		UserID:  *profile.OwnerUserID,
		Type:    notificationType,
		Payload: payloadJSON,
	})
	if err != nil {
		logger.Errorf(ctx, "failed to store a %s notification for user %s: %v", notificationType, *profile.OwnerUserID, err)
	}
}

func (s *notificationService) NotifyProfileMerged(ctx context.Context, queries sql.Queries, profile *domain.Profile, sourceUsername string) {
	if profile == nil || profile.OwnerUserID == nil {
		return
	}

	payload := domain.ProfileMergeNotificationPayload{
		ProfileID:       profile.ID,
		ProfileUsername: profile.MinecraftUsername,
		SourceUsername:  sourceUsername,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf(ctx, "failed to build the payload of a profile merge notification: %v", err)
		return
	}

	err = queries.InsertNotification(ctx, sql.InsertNotificationParams{
		UserID:  *profile.OwnerUserID,
		Type:    domain.NotificationTypeProfileMerged,
		Payload: payloadJSON,
	})
	if err != nil {
		logger.Errorf(ctx, "failed to store a profile merge notification for user %s: %v", *profile.OwnerUserID, err)
	}
}

// fillCosmetic copies the name and the look the item has right now, so the notification keeps them even after the item changes.
func (s *notificationService) fillCosmetic(ctx context.Context, queries sql.Queries, payload *domain.CosmeticNotificationPayload) error {
	switch payload.GrantType {
	case domain.GrantTypeNameColor:
		nameColor, err := queries.FindNameColorByID(ctx, payload.ItemID)
		if err != nil {
			return err
		}
		payload.ItemName = nameColor.Name
		payload.Colors = nameColor.Metadata.Colors
		return nil
	case domain.GrantTypeNamePrefix:
		namePrefix, err := queries.FindNamePrefixByID(ctx, payload.ItemID)
		if err != nil {
			return err
		}
		payload.ItemName = namePrefix.Name
		payload.PrefixImage = namePrefix.Metadata.Image
		return nil
	}
	return utils.NewBadRequestError("cosmetic notifications cover a name color and a name prefix only", nil)
}
