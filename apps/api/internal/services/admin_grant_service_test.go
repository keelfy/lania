package services

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type fakeGrantQueries struct {
	sql.Queries
	grants           []*domain.Grant
	colorOptions     []*domain.ProfileNameColorOption
	prefixOptions    []*domain.ProfileNamePrefixOption
	selectedPrefixes []*domain.ProfilePrefix
	revokeCalls      int
	colorUpdates     []uuid.UUID
	deletedPrefixes  []domain.ProfilePrefixType
}

func (q *fakeGrantQueries) FindProfileGrants(context.Context, uuid.UUID, uuid.UUID) ([]*domain.Grant, error) {
	return q.grants, nil
}

func (q *fakeGrantQueries) revoke(id uuid.UUID) {
	q.revokeCalls++
	now := time.Now()
	for _, grant := range q.grants {
		if grant.ID == id {
			grant.RevokedAt = &now
		}
	}
}

func (q *fakeGrantQueries) RevokeProfileAccess(_ context.Context, _, id uuid.UUID, _ *uuid.UUID) error {
	q.revoke(id)
	return nil
}

func (q *fakeGrantQueries) RevokeProfileNameColorOption(_ context.Context, _, id uuid.UUID, _ *uuid.UUID) error {
	q.revoke(id)
	return nil
}

func (q *fakeGrantQueries) RevokeProfileNamePrefixOption(_ context.Context, _, id uuid.UUID, _ *uuid.UUID) error {
	q.revoke(id)
	return nil
}

func (q *fakeGrantQueries) FindProfileNameColorOptionsByProfileID(context.Context, uuid.UUID, *uuid.UUID) ([]*domain.ProfileNameColorOption, error) {
	return q.colorOptions, nil
}

func (q *fakeGrantQueries) FindProfileNamePrefixOptionsByProfileIDAndType(context.Context, uuid.UUID, domain.ProfilePrefixType, *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error) {
	return q.prefixOptions, nil
}

func (q *fakeGrantQueries) FindProfilePrefixesByProfileID(context.Context, uuid.UUID) ([]*domain.ProfilePrefix, error) {
	return q.selectedPrefixes, nil
}

func (q *fakeGrantQueries) FindProfilePrefixesByProfileIDs(context.Context, uuid.UUIDs) ([]*domain.ProfilePrefix, error) {
	return q.selectedPrefixes, nil
}

func (q *fakeGrantQueries) UpdateProfileNameColorByID(_ context.Context, _ uuid.UUID, nameColorID uuid.UUID) error {
	q.colorUpdates = append(q.colorUpdates, nameColorID)
	return nil
}

func (q *fakeGrantQueries) DeleteProfilePrefixByProfileIDAndType(_ context.Context, _ uuid.UUID, prefixType domain.ProfilePrefixType) error {
	q.deletedPrefixes = append(q.deletedPrefixes, prefixType)
	return nil
}

type fakeGrantStorage struct {
	storage.MainStorage
	queries *fakeGrantQueries
	txCalls int
}

func (s *fakeGrantStorage) Queries() sql.Queries { return s.queries }

func (s *fakeGrantStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	s.txCalls++
	return fn(s.queries)
}

type stubProductService struct {
	ProductService
	products map[uuid.UUID]*domain.Product
}

func (s *stubProductService) GetProductByID(_ context.Context, id uuid.UUID) (*domain.Product, error) {
	if product, ok := s.products[id]; ok {
		return product, nil
	}
	return nil, utils.NewNotFoundError("product not found", nil)
}

type stubSeasonService struct {
	SeasonService
	known uuid.UUID
}

func (s *stubSeasonService) GetSeasonByID(_ context.Context, id uuid.UUID) (*domain.Season, error) {
	if id == s.known {
		return &domain.Season{}, nil
	}
	return nil, utils.NewNotFoundError("season not found", nil)
}

type grantCall struct {
	profileID, seasonID uuid.UUID
	product             *domain.Product
	source              domain.AccessSource
	orderItemID         *uuid.UUID
}

type stubFulfillmentService struct {
	FulfillmentService
	calls []grantCall
}

func (s *stubFulfillmentService) GrantProduct(_ context.Context, _ sql.Queries, profileID, seasonID uuid.UUID, product *domain.Product, source domain.AccessSource, orderItemID *uuid.UUID) error {
	s.calls = append(s.calls, grantCall{profileID, seasonID, product, source, orderItemID})
	return nil
}

type stubAccessService struct {
	AccessService
	hasAccess bool
}

func (s *stubAccessService) CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return s.hasAccess, nil
}

type recordingMinecraftService struct {
	MinecraftService
	removed  int
	prefixes []string
	err      error
}

func (s *recordingMinecraftService) RemoveFromWhitelist(context.Context, *domain.Profile) error {
	s.removed++
	return s.err
}

func (s *recordingMinecraftService) SetPrefixByMinecraftUUID(_ context.Context, _ uuid.UUID, prefix string) error {
	s.prefixes = append(s.prefixes, prefix)
	return s.err
}

type grantFixture struct {
	svc         AdminGrantService
	queries     *fakeGrantQueries
	storage     *fakeGrantStorage
	fulfillment *stubFulfillmentService
	access      *stubAccessService
	minecraft   *recordingMinecraftService
	profile     *domain.Profile
	season      uuid.UUID
	defaultDye  uuid.UUID
}

func newGrantFixture(t *testing.T) *grantFixture {
	t.Helper()

	f := &grantFixture{
		queries:     &fakeGrantQueries{},
		fulfillment: &stubFulfillmentService{},
		access:      &stubAccessService{},
		minecraft:   &recordingMinecraftService{},
		season:      uuid.New(),
		defaultDye:  uuid.New(),
		profile:     &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), NameColorID: uuid.New()},
	}
	t.Setenv("ACTIVE_SEASON_ID", f.season.String())
	t.Setenv("DEFAULT_NAME_COLOR_ID", f.defaultDye.String())

	f.storage = &fakeGrantStorage{queries: f.queries}
	profiles := &stubProfileService{profiles: map[uuid.UUID]*domain.Profile{f.profile.ID: f.profile}}
	f.svc = NewAdminGrantService(
		f.storage,
		profiles,
		&stubProductService{products: map[uuid.UUID]*domain.Product{}},
		&stubSeasonService{known: f.season},
		f.fulfillment,
		f.access,
		NewProfileCosmeticsService(f.storage),
		f.minecraft,
	)
	return f
}

func (f *grantFixture) addProduct(t *testing.T, category domain.ProductCategory, metadata any) *domain.Product {
	t.Helper()

	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	product := &domain.Product{ID: uuid.New(), Category: category, Metadata: raw}
	f.svc.(*adminGrantService).productService.(*stubProductService).products[product.ID] = product
	return product
}

func statusOf(err error) int {
	if err == nil {
		return http.StatusOK
	}
	return utils.MapCustomErrorToHttpStatus(err)
}

func TestGrantProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("hands the product over as an admin grant inside a transaction", func(t *testing.T) {
		f := newGrantFixture(t)
		product := f.addProduct(t, domain.ProductCategoryUpgrade, domain.UpgradeProductMetadata{Action: domain.ProductUpgradeActionSeasonAccess})

		if err := f.svc.GrantProduct(ctx, f.profile.ID, product.ID, f.season); err != nil {
			t.Fatal(err)
		}

		if f.storage.txCalls != 1 || len(f.fulfillment.calls) != 1 {
			t.Fatalf("got %d transactions and %d grants, want 1 and 1", f.storage.txCalls, len(f.fulfillment.calls))
		}
		call := f.fulfillment.calls[0]
		if call.profileID != f.profile.ID || call.seasonID != f.season || call.product != product ||
			call.source != domain.AccessSourceAdmin || call.orderItemID != nil {
			t.Errorf("got %+v", call)
		}
	})

	t.Run("refuses a product the profile already has", func(t *testing.T) {
		f := newGrantFixture(t)
		f.access.hasAccess = true
		product := f.addProduct(t, domain.ProductCategoryUpgrade, domain.UpgradeProductMetadata{Action: domain.ProductUpgradeActionSeasonAccess})

		err := f.svc.GrantProduct(ctx, f.profile.ID, product.ID, f.season)
		if statusOf(err) != http.StatusConflict || len(f.fulfillment.calls) != 0 {
			t.Fatalf("got status %d and %d grants, want conflict and none", statusOf(err), len(f.fulfillment.calls))
		}
	})

	t.Run("refuses a name color the profile already has", func(t *testing.T) {
		f := newGrantFixture(t)
		colorID := uuid.New()
		f.queries.colorOptions = []*domain.ProfileNameColorOption{{NameColorID: colorID}}
		product := f.addProduct(t, domain.ProductCategoryNameColor, domain.NameColorProductMetadata{NameColorID: colorID})

		if err := f.svc.GrantProduct(ctx, f.profile.ID, product.ID, f.season); statusOf(err) != http.StatusConflict {
			t.Fatalf("got status %d, want conflict", statusOf(err))
		}
	})

	t.Run("grants a name prefix the profile lacks", func(t *testing.T) {
		f := newGrantFixture(t)
		f.queries.prefixOptions = []*domain.ProfileNamePrefixOption{{NamePrefixID: uuid.New()}}
		product := f.addProduct(t, domain.ProductCategoryNamePrefix, domain.NamePrefixProductMetadata{NamePrefixID: uuid.New()})

		if err := f.svc.GrantProduct(ctx, f.profile.ID, product.ID, f.season); err != nil || len(f.fulfillment.calls) != 1 {
			t.Fatalf("got error %v and %d grants, want one grant", err, len(f.fulfillment.calls))
		}
	})

	t.Run("reports unknown profile, season and product", func(t *testing.T) {
		f := newGrantFixture(t)
		product := f.addProduct(t, domain.ProductCategoryUpgrade, domain.UpgradeProductMetadata{Action: domain.ProductUpgradeActionSeasonAccess})

		for name, err := range map[string]error{
			"profile": f.svc.GrantProduct(ctx, uuid.New(), product.ID, f.season),
			"season":  f.svc.GrantProduct(ctx, f.profile.ID, product.ID, uuid.New()),
			"product": f.svc.GrantProduct(ctx, f.profile.ID, uuid.New(), f.season),
		} {
			if statusOf(err) != http.StatusNotFound {
				t.Errorf("unknown %s: got status %d, want not found", name, statusOf(err))
			}
		}
		if len(f.fulfillment.calls) != 0 {
			t.Errorf("got %d grants, want none", len(f.fulfillment.calls))
		}
	})
}

func TestRevokeAccess(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		seasonIsLive  bool
		otherAccess   bool
		alreadyGone   bool
		wantRevokes   int
		wantWhitelist int
	}{
		{"removes the player from the whitelist", true, false, false, 1, 1},
		{"keeps the whitelist while another access remains", true, true, false, 1, 0},
		{"leaves the whitelist alone for a past season", false, false, false, 1, 0},
		{"repeating a revoke retries the whitelist update only", true, false, true, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newGrantFixture(t)
			seasonID := uuid.New()
			if tt.seasonIsLive {
				seasonID = f.season
			}
			grant := &domain.Grant{ID: uuid.New(), Type: domain.GrantTypeAccess, SeasonID: &seasonID}
			if tt.alreadyGone {
				now := time.Now()
				grant.RevokedAt = &now
			}
			f.queries.grants = []*domain.Grant{grant}
			f.access.hasAccess = tt.otherAccess

			if err := f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeAccess, grant.ID); err != nil {
				t.Fatal(err)
			}
			if f.queries.revokeCalls != tt.wantRevokes || f.minecraft.removed != tt.wantWhitelist {
				t.Errorf("got %d revokes and %d whitelist removals, want %d and %d",
					f.queries.revokeCalls, f.minecraft.removed, tt.wantRevokes, tt.wantWhitelist)
			}
		})
	}

	t.Run("reports a failed server update after the database is settled", func(t *testing.T) {
		f := newGrantFixture(t)
		grant := &domain.Grant{ID: uuid.New(), Type: domain.GrantTypeAccess, SeasonID: &f.season}
		f.queries.grants = []*domain.Grant{grant}
		f.minecraft.err = context.DeadlineExceeded

		err := f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeAccess, grant.ID)
		if statusOf(err) != http.StatusInternalServerError || !grant.IsRevoked() {
			t.Fatalf("got status %d and revoked %v, want an error and a revoked grant", statusOf(err), grant.IsRevoked())
		}
	})

	t.Run("unknown grant and mismatching type are not found", func(t *testing.T) {
		f := newGrantFixture(t)
		grant := &domain.Grant{ID: uuid.New(), Type: domain.GrantTypeAccess, SeasonID: &f.season}
		f.queries.grants = []*domain.Grant{grant}

		for name, err := range map[string]error{
			"id":   f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeAccess, uuid.New()),
			"type": f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeNameColor, grant.ID),
		} {
			if statusOf(err) != http.StatusNotFound {
				t.Errorf("unknown %s: got status %d, want not found", name, statusOf(err))
			}
		}
		if f.queries.revokeCalls != 0 {
			t.Errorf("got %d revokes, want none", f.queries.revokeCalls)
		}
	})
}

func TestRevokeNameColor(t *testing.T) {
	ctx := context.Background()

	t.Run("resets the selected color when nothing else allows it", func(t *testing.T) {
		f := newGrantFixture(t)
		grant := &domain.Grant{ID: uuid.New(), Type: domain.GrantTypeNameColor, ItemID: f.profile.NameColorID, SeasonID: &f.season}
		f.queries.grants = []*domain.Grant{grant}

		if err := f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeNameColor, grant.ID); err != nil {
			t.Fatal(err)
		}
		if len(f.queries.colorUpdates) != 1 || f.queries.colorUpdates[0] != f.defaultDye {
			t.Errorf("got color updates %v, want the default color", f.queries.colorUpdates)
		}
		if len(f.minecraft.prefixes) != 1 {
			t.Errorf("got %d prefix updates, want 1", len(f.minecraft.prefixes))
		}
	})

	t.Run("keeps the selected color while another option allows it", func(t *testing.T) {
		f := newGrantFixture(t)
		grant := &domain.Grant{ID: uuid.New(), Type: domain.GrantTypeNameColor, ItemID: f.profile.NameColorID, SeasonID: &f.season}
		f.queries.grants = []*domain.Grant{grant}
		f.queries.colorOptions = []*domain.ProfileNameColorOption{{NameColorID: f.profile.NameColorID}}

		if err := f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeNameColor, grant.ID); err != nil {
			t.Fatal(err)
		}
		if len(f.queries.colorUpdates) != 0 {
			t.Errorf("got color updates %v, want none", f.queries.colorUpdates)
		}
	})

	t.Run("keeps a selected color that is not the revoked one", func(t *testing.T) {
		f := newGrantFixture(t)
		grant := &domain.Grant{ID: uuid.New(), Type: domain.GrantTypeNameColor, ItemID: uuid.New(), SeasonID: &f.season}
		f.queries.grants = []*domain.Grant{grant}

		if err := f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeNameColor, grant.ID); err != nil {
			t.Fatal(err)
		}
		if len(f.queries.colorUpdates) != 0 {
			t.Errorf("got color updates %v, want none", f.queries.colorUpdates)
		}
	})
}

func TestRevokeNamePrefix(t *testing.T) {
	ctx := context.Background()
	prefixID := uuid.New()

	t.Run("clears the selected prefix when nothing else allows it", func(t *testing.T) {
		f := newGrantFixture(t)
		grant := &domain.Grant{ID: uuid.New(), Type: domain.GrantTypeNamePrefix, ItemID: prefixID, PrefixType: domain.ProfilePrefixTypeGlyth, SeasonID: &f.season}
		f.queries.grants = []*domain.Grant{grant}
		f.queries.selectedPrefixes = []*domain.ProfilePrefix{{ProfileID: f.profile.ID, NamePrefixID: prefixID, Type: domain.ProfilePrefixTypeGlyth}}

		if err := f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeNamePrefix, grant.ID); err != nil {
			t.Fatal(err)
		}
		if len(f.queries.deletedPrefixes) != 1 || f.queries.deletedPrefixes[0] != domain.ProfilePrefixTypeGlyth {
			t.Errorf("got cleared prefixes %v, want the glyth prefix", f.queries.deletedPrefixes)
		}
		if len(f.minecraft.prefixes) != 1 {
			t.Errorf("got %d prefix updates, want 1", len(f.minecraft.prefixes))
		}
	})

	t.Run("keeps a prefix of the same type that is not the revoked one", func(t *testing.T) {
		f := newGrantFixture(t)
		grant := &domain.Grant{ID: uuid.New(), Type: domain.GrantTypeNamePrefix, ItemID: prefixID, PrefixType: domain.ProfilePrefixTypeGlyth, SeasonID: &f.season}
		f.queries.grants = []*domain.Grant{grant}
		f.queries.selectedPrefixes = []*domain.ProfilePrefix{{ProfileID: f.profile.ID, NamePrefixID: uuid.New(), Type: domain.ProfilePrefixTypeGlyth}}

		if err := f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeNamePrefix, grant.ID); err != nil {
			t.Fatal(err)
		}
		if len(f.queries.deletedPrefixes) != 0 {
			t.Errorf("got cleared prefixes %v, want none", f.queries.deletedPrefixes)
		}
	})

	t.Run("keeps the selected prefix while another option allows it", func(t *testing.T) {
		f := newGrantFixture(t)
		grant := &domain.Grant{ID: uuid.New(), Type: domain.GrantTypeNamePrefix, ItemID: prefixID, PrefixType: domain.ProfilePrefixTypeGlyth, SeasonID: &f.season}
		f.queries.grants = []*domain.Grant{grant}
		f.queries.selectedPrefixes = []*domain.ProfilePrefix{{ProfileID: f.profile.ID, NamePrefixID: prefixID, Type: domain.ProfilePrefixTypeGlyth}}
		f.queries.prefixOptions = []*domain.ProfileNamePrefixOption{{NamePrefixID: prefixID}}

		if err := f.svc.RevokeGrant(ctx, f.profile.ID, domain.GrantTypeNamePrefix, grant.ID); err != nil {
			t.Fatal(err)
		}
		if len(f.queries.deletedPrefixes) != 0 {
			t.Errorf("got cleared prefixes %v, want none", f.queries.deletedPrefixes)
		}
	})
}
