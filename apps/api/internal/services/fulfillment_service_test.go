package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

type recordingAccessService struct {
	AccessService
	seasonID uuid.UUID
	profile  *domain.Profile
	source   domain.AccessSource
}

func (s *recordingAccessService) ObtainAccessForProfile(_ context.Context, seasonID uuid.UUID, profile *domain.Profile, source domain.AccessSource) error {
	s.seasonID, s.profile, s.source = seasonID, profile, source
	return nil
}

type cosmeticOptionCall struct {
	profileID, itemID uuid.UUID
	seasonID          *uuid.UUID
	orderItemID       *uuid.UUID
}

type recordingCosmeticsService struct {
	ProfileCosmeticsService
	color, prefix *cosmeticOptionCall
}

func (s *recordingCosmeticsService) AddProfileNameColorOption(_ context.Context, _ sql.Queries, profileID, nameColorID uuid.UUID, forSeasonID, orderItemID *uuid.UUID) error {
	s.color = &cosmeticOptionCall{profileID, nameColorID, forSeasonID, orderItemID}
	return nil
}

func (s *recordingCosmeticsService) AddProfileNameGlythOption(_ context.Context, _ sql.Queries, profileID, namePrefixID uuid.UUID, forSeasonID, orderItemID *uuid.UUID) error {
	s.prefix = &cosmeticOptionCall{profileID, namePrefixID, forSeasonID, orderItemID}
	return nil
}

func TestFulfillmentGrantProduct(t *testing.T) {
	ctx := context.Background()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New()}
	seasonID := uuid.New()
	orderItemID := uuid.New()

	newService := func() (FulfillmentService, *recordingAccessService, *recordingCosmeticsService) {
		access := &recordingAccessService{}
		cosmetics := &recordingCosmeticsService{}
		profiles := &stubProfileService{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}}
		return NewFulfillmentService(access, cosmetics, profiles), access, cosmetics
	}
	product := func(category domain.ProductCategory, metadata any) *domain.Product {
		raw, _ := json.Marshal(metadata)
		return &domain.Product{ID: uuid.New(), Category: category, Metadata: raw}
	}

	t.Run("season access passes the source on", func(t *testing.T) {
		svc, access, _ := newService()
		upgrade := product(domain.ProductCategoryUpgrade, domain.UpgradeProductMetadata{Action: domain.ProductUpgradeActionSeasonAccess})

		if err := svc.GrantProduct(ctx, nil, profile.ID, seasonID, upgrade, domain.AccessSourceAdmin, nil); err != nil {
			t.Fatal(err)
		}
		if access.profile == nil || access.profile.ID != profile.ID || access.seasonID != seasonID || access.source != domain.AccessSourceAdmin {
			t.Errorf("got %+v", access)
		}
	})

	t.Run("name color keeps the order item", func(t *testing.T) {
		svc, _, cosmetics := newService()
		colorID := uuid.New()
		color := product(domain.ProductCategoryNameColor, domain.NameColorProductMetadata{NameColorID: colorID})

		if err := svc.GrantProduct(ctx, nil, profile.ID, seasonID, color, domain.AccessSourceFreekassa, &orderItemID); err != nil {
			t.Fatal(err)
		}
		got := cosmetics.color
		if got == nil || got.profileID != profile.ID || got.itemID != colorID || *got.seasonID != seasonID || *got.orderItemID != orderItemID {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("name prefix without an order has no order item", func(t *testing.T) {
		svc, _, cosmetics := newService()
		prefixID := uuid.New()
		prefix := product(domain.ProductCategoryNamePrefix, domain.NamePrefixProductMetadata{NamePrefixID: prefixID})

		if err := svc.GrantProduct(ctx, nil, profile.ID, seasonID, prefix, domain.AccessSourceAdmin, nil); err != nil {
			t.Fatal(err)
		}
		got := cosmetics.prefix
		if got == nil || got.itemID != prefixID || got.orderItemID != nil {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("unknown category and unknown upgrade action fail", func(t *testing.T) {
		svc, _, _ := newService()

		if err := svc.GrantProduct(ctx, nil, profile.ID, seasonID, product("mystery", struct{}{}), domain.AccessSourceAdmin, nil); err == nil {
			t.Error("unknown category: got no error")
		}
		if err := svc.GrantProduct(ctx, nil, profile.ID, seasonID, product(domain.ProductCategoryUpgrade, domain.UpgradeProductMetadata{Action: "teleport"}), domain.AccessSourceAdmin, nil); err == nil {
			t.Error("unknown action: got no error")
		}
	})
}
