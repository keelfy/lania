package services

import (
	"context"
	stdsql "database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

// streamerQueries keeps profiles, streamers and applications in memory.
type streamerQueries struct {
	sql.Queries
	profiles     map[uuid.UUID]*domain.Profile
	streamers    map[uuid.UUID]*domain.Streamer
	applications []*domain.StreamerApplication
	// touched are the profiles whose role was stamped for role sync.
	touched uuid.UUIDs
}

func (q *streamerQueries) LockProfile(context.Context, uuid.UUID) error { return nil }

func (q *streamerQueries) FindProfileByID(_ context.Context, profileID uuid.UUID) (*domain.Profile, error) {
	profile, ok := q.profiles[profileID]
	if !ok {
		return nil, stdsql.ErrNoRows
	}
	return profile, nil
}

func (q *streamerQueries) FindStreamer(_ context.Context, profileID uuid.UUID) (*domain.Streamer, error) {
	streamer, ok := q.streamers[profileID]
	if !ok {
		return nil, stdsql.ErrNoRows
	}
	return streamer, nil
}

func (q *streamerQueries) SaveStreamer(_ context.Context, arg sql.SaveStreamerParams) error {
	q.streamers[arg.ProfileID] = &domain.Streamer{ProfileID: arg.ProfileID, Channels: arg.Channels, Description: arg.Description}
	return nil
}

func (q *streamerQueries) DeleteStreamer(_ context.Context, profileID uuid.UUID) (bool, error) {
	_, ok := q.streamers[profileID]
	delete(q.streamers, profileID)
	return ok, nil
}

func (q *streamerQueries) TouchProfileRole(_ context.Context, profileID uuid.UUID) error {
	q.touched = append(q.touched, profileID)
	return nil
}

func (q *streamerQueries) InsertStreamerApplication(_ context.Context, application *domain.StreamerApplication) error {
	q.applications = append(q.applications, application)
	return nil
}

func (q *streamerQueries) FindLatestStreamerApplication(_ context.Context, profileID uuid.UUID) (*domain.StreamerApplication, error) {
	for i := len(q.applications) - 1; i >= 0; i-- {
		if q.applications[i].ProfileID == profileID {
			return q.applications[i], nil
		}
	}
	return nil, stdsql.ErrNoRows
}

func (q *streamerQueries) FindStreamerApplicationByID(_ context.Context, applicationID uuid.UUID) (*domain.StreamerApplication, error) {
	for _, application := range q.applications {
		if application.ID == applicationID {
			return application, nil
		}
	}
	return nil, stdsql.ErrNoRows
}

func (q *streamerQueries) ReviewStreamerApplication(_ context.Context, applicationID uuid.UUID, status domain.StreamerApplicationStatus, reason *string, reviewedBy uuid.UUID) (bool, error) {
	for _, application := range q.applications {
		if application.ID == applicationID && application.Status == domain.StreamerApplicationStatusPending {
			now := time.Now()
			application.Status, application.RejectReason, application.ReviewedBy, application.ReviewedAt = status, reason, &reviewedBy, &now
			return true, nil
		}
	}
	return false, nil
}

type streamerStorage struct {
	storage.MainStorage
	queries *streamerQueries
}

func (s *streamerStorage) Queries() sql.Queries { return s.queries }

func (s *streamerStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	return fn(s.queries)
}

type streamerNotification struct {
	profileID        uuid.UUID
	notificationType domain.NotificationType
	reason           string
}

type recordingStreamerNotifications struct {
	NotificationService
	sent []streamerNotification
}

func (s *recordingStreamerNotifications) NotifyStreamer(_ context.Context, _ sql.Queries, profile *domain.Profile, notificationType domain.NotificationType, reason string) {
	s.sent = append(s.sent, streamerNotification{profile.ID, notificationType, reason})
}

type streamerFixture struct {
	svc           *streamerService
	queries       *streamerQueries
	notifications *recordingStreamerNotifications
	owner         uuid.UUID
	profile       *domain.Profile
	now           time.Time
}

func newStreamerFixture() *streamerFixture {
	owner := uuid.New()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "steve", OwnerUserID: &owner}
	f := &streamerFixture{
		queries: &streamerQueries{
			profiles:  map[uuid.UUID]*domain.Profile{profile.ID: profile},
			streamers: make(map[uuid.UUID]*domain.Streamer),
		},
		notifications: &recordingStreamerNotifications{},
		owner:         owner,
		profile:       profile,
		now:           time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC),
	}
	f.svc = NewStreamerService(&streamerStorage{queries: f.queries}, f.notifications).(*streamerService)
	f.svc.now = func() time.Time { return f.now }
	return f
}

var twitchChannel = []domain.StreamerChannel{{Platform: domain.StreamerPlatformTwitch, URL: " https://twitch.tv/steve "}}

func TestStreamerService_Apply(t *testing.T) {
	ctx := context.Background()

	t.Run("stores a pending application with trimmed input", func(t *testing.T) {
		f := newStreamerFixture()
		application, err := f.svc.Apply(ctx, f.owner, f.profile.ID, twitchChannel, "  I stream on Lania  ")
		if err != nil {
			t.Fatal(err)
		}
		if application.Status != domain.StreamerApplicationStatusPending || application.About != "I stream on Lania" ||
			application.Channels[0].URL != "https://twitch.tv/steve" || len(f.queries.applications) != 1 {
			t.Errorf("application = %+v, want a stored pending application with trimmed input", application)
		}
	})

	reviewedAt := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		prepare func(f *streamerFixture)
		userID  func(f *streamerFixture) uuid.UUID
		status  int
	}{
		{
			name:   "not the owner",
			userID: func(*streamerFixture) uuid.UUID { return uuid.New() },
			status: http.StatusForbidden,
		},
		{
			name: "already a streamer",
			prepare: func(f *streamerFixture) {
				f.queries.streamers[f.profile.ID] = &domain.Streamer{ProfileID: f.profile.ID}
			},
			status: http.StatusConflict,
		},
		{
			name: "an application under review",
			prepare: func(f *streamerFixture) {
				f.queries.applications = append(f.queries.applications, &domain.StreamerApplication{ID: uuid.New(), ProfileID: f.profile.ID, Status: domain.StreamerApplicationStatusPending})
			},
			status: http.StatusConflict,
		},
		{
			name: "rejected less than the cooldown ago",
			prepare: func(f *streamerFixture) {
				f.queries.applications = append(f.queries.applications, &domain.StreamerApplication{ID: uuid.New(), ProfileID: f.profile.ID, Status: domain.StreamerApplicationStatusRejected, ReviewedAt: &reviewedAt})
			},
			status: http.StatusTooManyRequests,
		},
		{
			name: "rejected longer than the cooldown ago",
			prepare: func(f *streamerFixture) {
				f.now = reviewedAt.Add(domain.StreamerReapplyCooldown)
				f.queries.applications = append(f.queries.applications, &domain.StreamerApplication{ID: uuid.New(), ProfileID: f.profile.ID, Status: domain.StreamerApplicationStatusRejected, ReviewedAt: &reviewedAt})
			},
			status: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newStreamerFixture()
			if tt.prepare != nil {
				tt.prepare(f)
			}
			userID := f.owner
			if tt.userID != nil {
				userID = tt.userID(f)
			}
			_, err := f.svc.Apply(ctx, userID, f.profile.ID, twitchChannel, "about")
			if got := statusOf(err); got != tt.status {
				t.Errorf("status = %d (%v), want %d", got, err, tt.status)
			}
		})
	}

	t.Run("bad input", func(t *testing.T) {
		f := newStreamerFixture()
		if _, err := f.svc.Apply(ctx, f.owner, f.profile.ID, nil, "about"); statusOf(err) != http.StatusBadRequest {
			t.Errorf("no channels error = %v, want bad request", err)
		}
		if _, err := f.svc.Apply(ctx, f.owner, f.profile.ID, twitchChannel, "   "); statusOf(err) != http.StatusBadRequest {
			t.Errorf("empty about error = %v, want bad request", err)
		}
	})
}

func TestStreamerService_Review(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()

	apply := func(t *testing.T, f *streamerFixture) *domain.StreamerApplication {
		t.Helper()
		application, err := f.svc.Apply(ctx, f.owner, f.profile.ID, twitchChannel, "about")
		if err != nil {
			t.Fatal(err)
		}
		return application
	}

	t.Run("approval makes a streamer, stamps the role and tells the owner", func(t *testing.T) {
		f := newStreamerFixture()
		application := apply(t, f)

		if err := f.svc.ApproveApplication(ctx, application.ID, adminID); err != nil {
			t.Fatal(err)
		}
		if streamer := f.queries.streamers[f.profile.ID]; streamer == nil || streamer.Channels[0].URL != "https://twitch.tv/steve" {
			t.Errorf("streamer = %+v, want the channels of the application", streamer)
		}
		if len(f.queries.touched) != 1 || application.Status != domain.StreamerApplicationStatusApproved {
			t.Errorf("touched %v, status %s; want the role stamped and the application approved", f.queries.touched, application.Status)
		}
		if len(f.notifications.sent) != 1 || f.notifications.sent[0].notificationType != domain.NotificationTypeStreamerApproved {
			t.Errorf("notifications = %v, want one approval", f.notifications.sent)
		}

		if err := f.svc.ApproveApplication(ctx, application.ID, adminID); statusOf(err) != http.StatusConflict {
			t.Errorf("second review error = %v, want conflict", err)
		}
	})

	t.Run("approval keeps a card an admin made meanwhile", func(t *testing.T) {
		f := newStreamerFixture()
		application := apply(t, f)
		description := "by admin"
		f.queries.streamers[f.profile.ID] = &domain.Streamer{ProfileID: f.profile.ID, Description: &description}

		if err := f.svc.ApproveApplication(ctx, application.ID, adminID); err != nil {
			t.Fatal(err)
		}
		if streamer := f.queries.streamers[f.profile.ID]; streamer.Description == nil || *streamer.Description != description {
			t.Errorf("streamer = %+v, want the admin's card", streamer)
		}
	})

	t.Run("rejection needs a reason and tells it to the owner", func(t *testing.T) {
		f := newStreamerFixture()
		application := apply(t, f)

		if err := f.svc.RejectApplication(ctx, application.ID, adminID, "  "); statusOf(err) != http.StatusBadRequest {
			t.Errorf("empty reason error = %v, want bad request", err)
		}
		if err := f.svc.RejectApplication(ctx, application.ID, adminID, "too few streams"); err != nil {
			t.Fatal(err)
		}
		if len(f.queries.streamers) != 0 || len(f.queries.touched) != 0 {
			t.Error("a rejection must not make a streamer")
		}
		if len(f.notifications.sent) != 1 || f.notifications.sent[0].reason != "too few streams" {
			t.Errorf("notifications = %v, want the rejection with its reason", f.notifications.sent)
		}
	})

	t.Run("an unknown application is not found", func(t *testing.T) {
		f := newStreamerFixture()
		if err := f.svc.ApproveApplication(ctx, uuid.New(), adminID); statusOf(err) != http.StatusNotFound {
			t.Errorf("error = %v, want not found", err)
		}
	})
}

func TestStreamerService_AdminGrantAndRevoke(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	f := newStreamerFixture()

	if _, err := f.svc.SaveStreamer(ctx, adminID, f.profile.ID, twitchChannel, ""); err != nil {
		t.Fatal(err)
	}
	if f.queries.streamers[f.profile.ID] == nil || len(f.queries.touched) != 1 {
		t.Fatalf("streamers %v, touched %v; want a streamer with the role stamped", f.queries.streamers, f.queries.touched)
	}

	if err := f.svc.RevokeStreamer(ctx, adminID, f.profile.ID); err != nil {
		t.Fatal(err)
	}
	if len(f.queries.streamers) != 0 || len(f.queries.touched) != 2 {
		t.Errorf("streamers %v, touched %v; want the role gone and stamped again", f.queries.streamers, f.queries.touched)
	}
	if len(f.notifications.sent) != 1 || f.notifications.sent[0].notificationType != domain.NotificationTypeStreamerRevoked {
		t.Errorf("notifications = %v, want one revocation", f.notifications.sent)
	}
	if err := f.svc.RevokeStreamer(ctx, adminID, f.profile.ID); statusOf(err) != http.StatusNotFound {
		t.Errorf("revoking a non-streamer error = %v, want not found", err)
	}
}

func TestStreamerService_UpdateOwnedStreamer(t *testing.T) {
	ctx := context.Background()
	f := newStreamerFixture()

	if _, err := f.svc.UpdateOwnedStreamer(ctx, f.owner, f.profile.ID, twitchChannel, "hi"); statusOf(err) != http.StatusForbidden {
		t.Errorf("a non-streamer edit error = %v, want forbidden", err)
	}

	f.queries.streamers[f.profile.ID] = &domain.Streamer{ProfileID: f.profile.ID}
	streamer, err := f.svc.UpdateOwnedStreamer(ctx, f.owner, f.profile.ID, twitchChannel, "  hi  ")
	if err != nil {
		t.Fatal(err)
	}
	if streamer.Description == nil || *streamer.Description != "hi" {
		t.Errorf("description = %v, want the trimmed text", streamer.Description)
	}
	if len(f.queries.touched) != 0 {
		t.Error("editing the card must not push the role again")
	}
}
