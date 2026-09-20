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
	inserted     []sql.InsertNotificationParams
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
			nameColors:   map[uuid.UUID]*domain.NameColor{colorID: {ID: colorID, Name: "Aurora"}},
			namePrefixes: map[uuid.UUID]*domain.NamePrefix{prefixID: {ID: prefixID, Name: "Star"}},
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
		if payload.SeasonID == nil || *payload.SeasonID != seasonID {
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
		if payload.PrefixType != domain.ProfilePrefixTypeSpecial || payload.ItemName != "Star" || payload.SeasonID != nil {
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
}
