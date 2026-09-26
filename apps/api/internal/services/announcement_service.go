package services

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

// announcementBatchSize bounds the rows of one insert, far below the placeholder limit of MariaDB.
const announcementBatchSize = 500

type AnnouncementService interface {
	// SendAnnouncement stores the announcement as a notification for every user and returns how many got it.
	// A user who signs up later does not get it.
	SendAnnouncement(ctx context.Context, payload domain.AnnouncementNotificationPayload) (int, error)
}

type announcementService struct {
	oryClient clients.OryAPI
	storage   storage.MainStorage
}

func NewAnnouncementService(oryClient clients.OryAPI, storage storage.MainStorage) AnnouncementService {
	return &announcementService{oryClient: oryClient, storage: storage}
}

func (s *announcementService) SendAnnouncement(ctx context.Context, payload domain.AnnouncementNotificationPayload) (int, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return 0, utils.NewInternalServerError("failed to build the payload of an announcement", err)
	}

	userIDs, err := s.listUserIDs(ctx)
	if err != nil {
		return 0, err
	}

	// One transaction, so a failed batch leaves nobody with the announcement and the admin can simply send it again.
	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		for batch := range slices.Chunk(userIDs, announcementBatchSize) {
			if err := queries.InsertNotifications(ctx, batch, domain.NotificationTypeAnnouncement, payloadJSON); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, utils.NewInternalServerError("failed to store the announcement", err)
	}
	return len(userIDs), nil
}

// listUserIDs walks every identity page by page, up to identityScanMaxPages.
func (s *announcementService) listUserIDs(ctx context.Context) (uuid.UUIDs, error) {
	userIDs := make(uuid.UUIDs, 0, identityScanPageSize)
	pageToken := ""

	for range identityScanMaxPages {
		identities, nextPageToken, err := s.oryClient.ListIdentities(ctx, identityScanPageSize, pageToken, "")
		if err != nil {
			return nil, utils.NewInternalServerError("failed to list identities", err)
		}
		for _, user := range usersFromIdentities(identities) {
			userIDs = append(userIDs, user.ID)
		}

		if nextPageToken == "" {
			return userIDs, nil
		}
		pageToken = nextPageToken
	}

	logger.Warnf(ctx, "identity scan stopped after %d pages, the announcement misses the users past them", identityScanMaxPages)
	return userIDs, nil
}
