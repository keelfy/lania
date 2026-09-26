package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	ory "github.com/ory/client-go"
)

type fakeAnnouncementQueries struct {
	sql.Queries
	batches []uuid.UUIDs
	payload json.RawMessage
}

func (q *fakeAnnouncementQueries) InsertNotifications(_ context.Context, userIDs uuid.UUIDs, notificationType domain.NotificationType, payload json.RawMessage) error {
	if notificationType != domain.NotificationTypeAnnouncement {
		return nil
	}
	q.batches = append(q.batches, userIDs)
	q.payload = payload
	return nil
}

type txMainStorage struct {
	storage.MainStorage
	queries sql.Queries
}

func (s *txMainStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	return fn(s.queries)
}

func TestSendAnnouncement(t *testing.T) {
	ctx := context.Background()

	identities := make([]ory.Identity, 0, 1201)
	for range 1200 {
		identities = append(identities, identity("", ""))
	}
	// An identity without a UUID is no user of the site.
	identities = append(identities, ory.Identity{Id: "not-a-uuid"})

	queries := &fakeAnnouncementQueries{}
	svc := NewAnnouncementService(&fakeOry{identities: identities, pageSize: 250}, &txMainStorage{queries: queries})

	payload := domain.AnnouncementNotificationPayload{
		Title: map[string]string{"ru": "Новые товары", "en": "New products"},
		Link:  "/shop",
	}
	recipients, err := svc.SendAnnouncement(ctx, payload)
	if err != nil {
		t.Fatal(err)
	}
	if recipients != 1200 {
		t.Errorf("recipients = %d, want 1200", recipients)
	}

	seen := make(map[uuid.UUID]bool)
	for _, batch := range queries.batches {
		if len(batch) > announcementBatchSize {
			t.Errorf("batch of %d rows is over %d", len(batch), announcementBatchSize)
		}
		for _, id := range batch {
			seen[id] = true
		}
	}
	if len(queries.batches) != 3 || len(seen) != 1200 {
		t.Errorf("got %d batches for %d distinct users, want 3 batches for 1200", len(queries.batches), len(seen))
	}

	var stored domain.AnnouncementNotificationPayload
	if err := json.Unmarshal(queries.payload, &stored); err != nil {
		t.Fatal(err)
	}
	if stored.Title["en"] != "New products" || stored.Link != "/shop" || stored.Body != nil {
		t.Errorf("stored payload = %+v", stored)
	}
}
