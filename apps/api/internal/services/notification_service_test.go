package services

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

type fakeNotificationQueries struct {
	sql.Queries
	nameColors   map[uuid.UUID]*domain.NameColor
	namePrefixes map[uuid.UUID]*domain.NamePrefix
	seasons      map[uuid.UUID]*domain.Season
	inserted     []sql.InsertNotificationParams
	// stored and gotFilter back FindNotificationsByUserID.
	stored    []*domain.Notification
	gotFilter domain.NotificationFilter
}

func (q *fakeNotificationQueries) FindSeasonByID(_ context.Context, id uuid.UUID) (*domain.Season, error) {
	if season, ok := q.seasons[id]; ok {
		return season, nil
	}
	return nil, stdsql.ErrNoRows
}

func (q *fakeNotificationQueries) FindNotificationsByUserID(_ context.Context, _ uuid.UUID, filter domain.NotificationFilter) ([]*domain.Notification, error) {
	q.gotFilter = filter
	end := min(filter.Offset+filter.Limit, len(q.stored))
	if filter.Offset >= end {
		return []*domain.Notification{}, nil
	}
	return q.stored[filter.Offset:end], nil
}

func (q *fakeNotificationQueries) FindNameColorByID(_ context.Context, id uuid.UUID) (*domain.NameColor, error) {
	if nameColor, ok := q.nameColors[id]; ok {
		return nameColor, nil
	}
	return nil, stdsql.ErrNoRows
}

func (q *fakeNotificationQueries) FindNamePrefixByID(_ context.Context, id uuid.UUID) (*domain.NamePrefix, error) {
	if namePrefix, ok := q.namePrefixes[id]; ok {
		return namePrefix, nil
	}
	return nil, stdsql.ErrNoRows
}

func (q *fakeNotificationQueries) InsertNotification(_ context.Context, arg sql.InsertNotificationParams) error {
	q.inserted = append(q.inserted, arg)
	return nil
}

func TestNotifyCosmeticGranted(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	colorID := uuid.New()
	prefixID := uuid.New()
	seasonID := uuid.New()

	newFixture := func() (NotificationService, *fakeNotificationQueries) {
		queries := &fakeNotificationQueries{
			nameColors:   map[uuid.UUID]*domain.NameColor{colorID: {ID: colorID, Name: "Aurora", Metadata: domain.NameColorMetadata{Colors: []string{"#7fd3ff", "#ffffff"}}}},
			namePrefixes: map[uuid.UUID]*domain.NamePrefix{prefixID: {ID: prefixID, Name: "Star", Metadata: domain.NamePrefixMetadata{Image: "/prefixes/star.png"}}},
			seasons:      map[uuid.UUID]*domain.Season{seasonID: {ID: seasonID, Name: "Season 2"}},
		}
		return NewNotificationService(nil), queries
	}
	owned := &domain.Profile{ID: uuid.New(), MinecraftUsername: "keelfy", OwnerUserID: &ownerID}

	t.Run("stores the name color for the owner of the profile", func(t *testing.T) {
		svc, queries := newFixture()

		svc.NotifyCosmeticGranted(ctx, queries, owned, domain.GrantTypeNameColor, colorID, "", &seasonID)

		if len(queries.inserted) != 1 {
			t.Fatalf("got %d notifications", len(queries.inserted))
		}
		got := queries.inserted[0]
		if got.UserID != ownerID || got.Type != domain.NotificationTypeCosmeticGranted {
			t.Errorf("got %+v", got)
		}

		var payload domain.CosmeticNotificationPayload
		if err := json.Unmarshal(got.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.ItemID != colorID || payload.ItemName != "Aurora" || payload.ProfileUsername != "keelfy" || payload.GrantType != domain.GrantTypeNameColor {
			t.Errorf("payload: got %+v", payload)
		}
		if len(payload.Colors) != 2 || payload.Colors[0] != "#7fd3ff" || payload.PrefixImage != "" {
			t.Errorf("payload look: got %+v", payload)
		}
		if payload.SeasonID == nil || *payload.SeasonID != seasonID || payload.SeasonName != "Season 2" {
			t.Errorf("payload season: got %+v", payload.SeasonID)
		}
	})

	t.Run("keeps the prefix type of a name prefix", func(t *testing.T) {
		svc, queries := newFixture()

		svc.NotifyCosmeticRevoked(ctx, queries, owned, domain.GrantTypeNamePrefix, prefixID, domain.ProfilePrefixTypeSpecial, nil)

		if len(queries.inserted) != 1 {
			t.Fatalf("got %d notifications", len(queries.inserted))
		}
		got := queries.inserted[0]
		if got.Type != domain.NotificationTypeCosmeticRevoked {
			t.Errorf("got type %s", got.Type)
		}

		var payload domain.CosmeticNotificationPayload
		if err := json.Unmarshal(got.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.PrefixType != domain.ProfilePrefixTypeSpecial || payload.ItemName != "Star" || payload.SeasonID != nil || payload.SeasonName != "" {
			t.Errorf("payload: got %+v", payload)
		}
		if payload.PrefixImage != "/prefixes/star.png" || payload.Colors != nil {
			t.Errorf("payload: got %+v", payload)
		}
	})

	t.Run("a profile without an owner and an unknown item store nothing", func(t *testing.T) {
		svc, queries := newFixture()

		svc.NotifyCosmeticGranted(ctx, queries, &domain.Profile{ID: uuid.New()}, domain.GrantTypeNameColor, colorID, "", nil)
		svc.NotifyCosmeticGranted(ctx, queries, owned, domain.GrantTypeNameColor, uuid.New(), "", nil)
		svc.NotifyCosmeticGranted(ctx, queries, owned, domain.GrantTypeAccess, colorID, "", nil)

		if len(queries.inserted) != 0 {
			t.Errorf("got %d notifications", len(queries.inserted))
		}
	})

	t.Run("an unknown season still stores the notification", func(t *testing.T) {
		svc, queries := newFixture()
		unknown := uuid.New()

		svc.NotifyCosmeticGranted(ctx, queries, owned, domain.GrantTypeNameColor, colorID, "", &unknown)

		if len(queries.inserted) != 1 {
			t.Fatalf("got %d notifications", len(queries.inserted))
		}
	})
}

func TestGetNotifications(t *testing.T) {
	ctx := context.Background()
	stored := make([]*domain.Notification, 5)
	for i := range stored {
		stored[i] = &domain.Notification{ID: uuid.New()}
	}

	tests := []struct {
		name        string
		filter      domain.NotificationFilter
		wantLen     int
		wantHasMore bool
	}{
		{"a page with more after it", domain.NotificationFilter{Limit: 2}, 2, true},
		{"the last full page", domain.NotificationFilter{Offset: 3, Limit: 2}, 2, false},
		{"a page past the end", domain.NotificationFilter{Offset: 10, Limit: 2}, 0, false},
		{"a negative offset starts at the top", domain.NotificationFilter{Offset: -4, Limit: 4}, 4, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := &fakeNotificationQueries{stored: stored}
			svc := NewNotificationService(&stubMainStorage{queries: queries})

			got, hasMore, err := svc.GetNotifications(ctx, uuid.New(), tt.filter)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tt.wantLen || hasMore != tt.wantHasMore {
				t.Errorf("got %d notifications, hasMore %v", len(got), hasMore)
			}
		})
	}

	t.Run("keeps the unread filter and caps the limit", func(t *testing.T) {
		queries := &fakeNotificationQueries{}
		svc := NewNotificationService(&stubMainStorage{queries: queries})

		if _, _, err := svc.GetNotifications(ctx, uuid.New(), domain.NotificationFilter{UnreadOnly: true, Limit: 1000}); err != nil {
			t.Fatal(err)
		}
		if !queries.gotFilter.UnreadOnly || queries.gotFilter.Limit != MaxNotifications+1 {
			t.Errorf("got filter %+v", queries.gotFilter)
		}
	})
}
