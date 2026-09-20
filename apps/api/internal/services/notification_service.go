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
	// GetNotifications returns the notifications of the user, newest first.
	GetNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Notification, error)
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
}

type notificationService struct {
	storage storage.MainStorage
}

func NewNotificationService(storage storage.MainStorage) NotificationService {
	return &notificationService{storage: storage}
}

func (s *notificationService) GetNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Notification, error) {
	if limit <= 0 || limit > MaxNotifications {
		limit = MaxNotifications
	}

	notifications, err := s.storage.Queries().FindNotificationsByUserID(ctx, userID, limit)
	if err == stdsql.ErrNoRows {
		return []*domain.Notification{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to find notifications", err)
	}
	return notifications, nil
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

	itemName, err := s.cosmeticName(ctx, queries, grantType, itemID)
	if err != nil {
		logger.Errorf(ctx, "failed to read the name of cosmetic %s for a notification: %v", itemID, err)
		return
	}

	payload, err := json.Marshal(domain.CosmeticNotificationPayload{
		ProfileID:       profile.ID,
		ProfileUsername: profile.MinecraftUsername,
		GrantType:       grantType,
		PrefixType:      prefixType,
		ItemID:          itemID,
		ItemName:        itemName,
		SeasonID:        seasonID,
	})
	if err != nil {
		logger.Errorf(ctx, "failed to build the payload of a cosmetic notification: %v", err)
		return
	}

	err = queries.InsertNotification(ctx, sql.InsertNotificationParams{
		UserID:  *profile.OwnerUserID,
		Type:    notificationType,
		Payload: payload,
	})
	if err != nil {
		logger.Errorf(ctx, "failed to store a %s notification for user %s: %v", notificationType, *profile.OwnerUserID, err)
	}
}

// cosmeticName returns the name the item has right now, so the notification keeps it even after a rename.
func (s *notificationService) cosmeticName(ctx context.Context, queries sql.Queries, grantType domain.GrantType, itemID uuid.UUID) (string, error) {
	switch grantType {
	case domain.GrantTypeNameColor:
		nameColor, err := queries.FindNameColorByID(ctx, itemID)
		if err != nil {
			return "", err
		}
		return nameColor.Name, nil
	case domain.GrantTypeNamePrefix:
		namePrefix, err := queries.FindNamePrefixByID(ctx, itemID)
		if err != nil {
			return "", err
		}
		return namePrefix.Name, nil
	}
	return "", utils.NewBadRequestError("cosmetic notifications cover a name color and a name prefix only", nil)
}
