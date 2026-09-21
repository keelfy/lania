package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

type selection struct {
	profileID, seasonID uuid.UUID
	color               uuid.UUID
	prefixType          domain.ProfilePrefixType
	prefix              *uuid.UUID
}

// cosmeticsQueries keeps the selection per (profile, season) like profile_season_cosmetics does.
type cosmeticsQueries struct {
	sql.Queries
	colors           map[[2]uuid.UUID]*domain.NameColor
	glyths           map[[2]uuid.UUID]*domain.NamePrefix
	writes           []selection
	insertedColors   []sql.InsertProfileNameColorOptionParams
	insertedPrefixes []sql.InsertProfileNamePrefixOptionParams
}

func (q *cosmeticsQueries) FindProfilesSeasonCosmetics(_ context.Context, profileIDs uuid.UUIDs, seasonID, defaultNameColorID uuid.UUID) (map[uuid.UUID]*domain.ProfileCosmetics, error) {
	cosmetics := make(map[uuid.UUID]*domain.ProfileCosmetics, len(profileIDs))
	for _, profileID := range profileIDs {
		key := [2]uuid.UUID{profileID, seasonID}
		selected := &domain.ProfileCosmetics{NameColor: q.colors[key], Glyth: q.glyths[key]}
		if selected.NameColor == nil {
			selected.NameColor = &domain.NameColor{ID: defaultNameColorID}
		}
		cosmetics[profileID] = selected
	}
	return cosmetics, nil
}

func (q *cosmeticsQueries) SetProfileSeasonNameColor(_ context.Context, profileID, seasonID, nameColorID uuid.UUID) error {
	q.writes = append(q.writes, selection{profileID: profileID, seasonID: seasonID, color: nameColorID})
	q.colors[[2]uuid.UUID{profileID, seasonID}] = &domain.NameColor{ID: nameColorID, Metadata: domain.NameColorMetadata{Colors: []string{nameColorID.String()}}}
	return nil
}

func (q *cosmeticsQueries) SetProfileSeasonPrefix(_ context.Context, profileID, seasonID uuid.UUID, prefixType domain.ProfilePrefixType, namePrefixID *uuid.UUID) error {
	q.writes = append(q.writes, selection{profileID: profileID, seasonID: seasonID, prefixType: prefixType, prefix: namePrefixID})
	key := [2]uuid.UUID{profileID, seasonID}
	if namePrefixID == nil {
		delete(q.glyths, key)
		return nil
	}
	q.glyths[key] = &domain.NamePrefix{ID: *namePrefixID, Metadata: domain.NamePrefixMetadata{Prefix: "[G]"}}
	return nil
}

func (q *cosmeticsQueries) InsertProfileNameColorOption(_ context.Context, arg sql.InsertProfileNameColorOptionParams) error {
	q.insertedColors = append(q.insertedColors, arg)
	return nil
}

func (q *cosmeticsQueries) InsertProfileNamePrefixOption(_ context.Context, arg sql.InsertProfileNamePrefixOptionParams) error {
	q.insertedPrefixes = append(q.insertedPrefixes, arg)
	return nil
}

type cosmeticsStorage struct {
	storage.MainStorage
	queries *cosmeticsQueries
}

func (s *cosmeticsStorage) Queries() sql.Queries { return s.queries }

func newCosmeticsFixture() (ProfileCosmeticsService, *cosmeticsQueries) {
	queries := &cosmeticsQueries{
		colors: make(map[[2]uuid.UUID]*domain.NameColor),
		glyths: make(map[[2]uuid.UUID]*domain.NamePrefix),
	}
	return NewProfileCosmeticsService(&cosmeticsStorage{queries: queries}), queries
}

func TestProfileCosmeticsService_SelectionBelongsToOneSeason(t *testing.T) {
	ctx := context.Background()
	t.Setenv("DEFAULT_NAME_COLOR_ID", uuid.New().String())
	service, queries := newCosmeticsFixture()
	profileID, seasonA, seasonB, color := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	if err := service.SelectProfileNameColor(ctx, queries, profileID, seasonA, color); err != nil {
		t.Fatal(err)
	}
	glyth := uuid.New()
	if err := service.SelectProfileNamePrefix(ctx, queries, profileID, seasonA, glyth, domain.ProfilePrefixTypeGlyth); err != nil {
		t.Fatal(err)
	}

	inA, err := service.GetProfileChatPrefix(ctx, profileID, seasonA)
	if err != nil {
		t.Fatal(err)
	}
	inB, err := service.GetProfileChatPrefix(ctx, profileID, seasonB)
	if err != nil {
		t.Fatal(err)
	}
	if inA != "<reset>[G] <color:"+color.String()+">" {
		t.Errorf("prefix in season A = %q, want the selected glyth and color", inA)
	}
	if inB != "<reset>" {
		t.Errorf("prefix in season B = %q, want nothing chosen there to show", inB)
	}

	if err := service.ClearProfilePrefixByType(ctx, queries, profileID, seasonA, domain.ProfilePrefixTypeGlyth); err != nil {
		t.Fatal(err)
	}
	if cleared, _ := service.GetProfileChatPrefix(ctx, profileID, seasonA); cleared != "<reset><color:"+color.String()+">" {
		t.Errorf("prefix in season A after clearing = %q, want the color only", cleared)
	}
	for _, write := range queries.writes {
		if write.seasonID != seasonA {
			t.Errorf("wrote a selection in season %s, want season A only", write.seasonID)
		}
	}
}

func TestProfileCosmeticsService_PurchasedOptionsAreNeverPermanent(t *testing.T) {
	ctx := context.Background()
	service, queries := newCosmeticsFixture()
	profileID, seasonID, orderItemID, itemID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	cases := map[string]func(season, orderItem *uuid.UUID) error{
		"name color": func(season, orderItem *uuid.UUID) error {
			return service.AddProfileNameColorOption(ctx, queries, profileID, itemID, season, orderItem)
		},
		"name prefix": func(season, orderItem *uuid.UUID) error {
			return service.AddProfileNameGlythOption(ctx, queries, profileID, itemID, season, orderItem)
		},
	}
	for name, add := range cases {
		t.Run(name, func(t *testing.T) {
			queries.insertedColors, queries.insertedPrefixes = nil, nil

			if err := add(nil, &orderItemID); statusOf(err) != http.StatusInternalServerError {
				t.Errorf("purchase without a season: status %d, want an error", statusOf(err))
			}
			if len(queries.insertedColors)+len(queries.insertedPrefixes) != 0 {
				t.Error("a purchase without a season was stored")
			}

			if err := add(&seasonID, &orderItemID); err != nil {
				t.Errorf("purchase for a season: %v", err)
			}
			if err := add(nil, nil); err != nil {
				t.Errorf("admin grant for good: %v", err)
			}
			if len(queries.insertedColors)+len(queries.insertedPrefixes) != 2 {
				t.Errorf("stored %d options, want the season purchase and the permanent grant", len(queries.insertedColors)+len(queries.insertedPrefixes))
			}
		})
	}
}
