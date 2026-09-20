package services

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type AdminGrantService interface {
	// GetProfileGrants returns everything the profile was given, revoked grants included, newest first.
	GetProfileGrants(ctx context.Context, profileID uuid.UUID) ([]*domain.Grant, error)
	// GrantProduct gives the product to the profile for the season. It fails with a conflict error
	// when the profile already has the product for that season.
	GrantProduct(ctx context.Context, profileID, productID, seasonID uuid.UUID) error
	// RevokeGrant takes the grant back and undoes its effect in the game.
	// Repeating the request for a revoked grant repeats the update of the Minecraft server,
	// so a failed update can be retried.
	RevokeGrant(ctx context.Context, profileID uuid.UUID, grantType domain.GrantType, grantID uuid.UUID) error
}

type adminGrantService struct {
	storage                 storage.MainStorage
	profileService          ProfileService
	productService          ProductService
	seasonService           SeasonService
	fulfillmentService      FulfillmentService
	accessService           AccessService
	profileCosmeticsService ProfileCosmeticsService
	minecraftService        MinecraftService
}

func NewAdminGrantService(
	storage storage.MainStorage,
	profileService ProfileService,
	productService ProductService,
	seasonService SeasonService,
	fulfillmentService FulfillmentService,
	accessService AccessService,
	profileCosmeticsService ProfileCosmeticsService,
	minecraftService MinecraftService,
) AdminGrantService {
	return &adminGrantService{
		storage:                 storage,
		profileService:          profileService,
		productService:          productService,
		seasonService:           seasonService,
		fulfillmentService:      fulfillmentService,
		accessService:           accessService,
		profileCosmeticsService: profileCosmeticsService,
		minecraftService:        minecraftService,
	}
}

func (s *adminGrantService) GetProfileGrants(ctx context.Context, profileID uuid.UUID) ([]*domain.Grant, error) {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	return s.findGrants(ctx, s.storage.Queries(), profile)
}

func (s *adminGrantService) GrantProduct(ctx context.Context, profileID, productID, seasonID uuid.UUID) error {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return err
	}
	if _, err := s.seasonService.GetSeasonByID(ctx, seasonID); err != nil {
		return err
	}
	product, err := s.productService.GetProductByID(ctx, productID)
	if err != nil {
		return err
	}

	hasProduct, err := s.profileHasProduct(ctx, profile, seasonID, product)
	if err != nil {
		return err
	}
	if hasProduct {
		return utils.NewConflictError("profile already has this product for the season", nil)
	}

	return s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		return s.fulfillmentService.GrantProduct(ctx, queries, profile.ID, seasonID, product, domain.AccessSourceAdmin, nil)
	})
}

func (s *adminGrantService) RevokeGrant(ctx context.Context, profileID uuid.UUID, grantType domain.GrantType, grantID uuid.UUID) error {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return err
	}

	grants, err := s.findGrants(ctx, s.storage.Queries(), profile)
	if err != nil {
		return err
	}
	index := slices.IndexFunc(grants, func(grant *domain.Grant) bool {
		return grant.Type == grantType && grant.ID == grantID
	})
	if index < 0 {
		return utils.NewNotFoundError("grant not found", nil)
	}
	grant := grants[index]

	if !grant.IsRevoked() {
		err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
			return s.revoke(ctx, queries, profile, grant)
		})
		if err != nil {
			return err
		}
	}

	// The database is settled by now, so a failure here leaves the grant revoked and can be retried.
	if err := s.updateGame(ctx, profile, grant); err != nil {
		return utils.NewInternalServerError("grant is revoked, but the Minecraft server was not updated, repeat the request to retry", err)
	}
	return nil
}

func (s *adminGrantService) findGrants(ctx context.Context, queries sql.Queries, profile *domain.Profile) ([]*domain.Grant, error) {
	grants, err := queries.FindProfileGrants(ctx, profile.ID, profile.MinecraftUUID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to find profile grants", err)
	}
	return grants, nil
}

func (s *adminGrantService) profileHasProduct(ctx context.Context, profile *domain.Profile, seasonID uuid.UUID, product *domain.Product) (bool, error) {
	switch product.Category {
	case domain.ProductCategoryUpgrade:
		var metadata domain.UpgradeProductMetadata
		if err := json.Unmarshal(product.Metadata, &metadata); err != nil {
			return false, utils.NewInternalServerError("failed to unmarshal upgrade product metadata", err)
		}
		if metadata.Action != domain.ProductUpgradeActionSeasonAccess {
			return false, utils.NewBadRequestError("product has an unknown upgrade action", nil)
		}
		return s.accessService.CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, profile.MinecraftUUID, seasonID)
	case domain.ProductCategoryNameColor:
		var metadata domain.NameColorProductMetadata
		if err := json.Unmarshal(product.Metadata, &metadata); err != nil {
			return false, utils.NewInternalServerError("failed to unmarshal name color product metadata", err)
		}
		options, err := s.profileCosmeticsService.GetProfileNameColorOptions(ctx, profile.ID, &seasonID)
		if err != nil {
			return false, err
		}
		return slices.ContainsFunc(options, func(option *domain.ProfileNameColorOption) bool {
			return option.NameColorID == metadata.NameColorID
		}), nil
	case domain.ProductCategoryNamePrefix:
		var metadata domain.NamePrefixProductMetadata
		if err := json.Unmarshal(product.Metadata, &metadata); err != nil {
			return false, utils.NewInternalServerError("failed to unmarshal name prefix product metadata", err)
		}
		options, err := s.profileCosmeticsService.GetProfileNamePrefixOptionsByProfileIDAndType(ctx, profile.ID, domain.ProfilePrefixTypeGlyth, &seasonID)
		if err != nil {
			return false, err
		}
		return slices.ContainsFunc(options, func(option *domain.ProfileNamePrefixOption) bool {
			return option.NamePrefixID == metadata.NamePrefixID
		}), nil
	}
	return false, utils.NewBadRequestError("product has an unknown category", nil)
}

// revoke marks the grant revoked and drops the selection of a name color or a name prefix that the profile can no longer use.
func (s *adminGrantService) revoke(ctx context.Context, queries sql.Queries, profile *domain.Profile, grant *domain.Grant) error {
	revokedBy := utils.GetUserIDFromContextOrNil(ctx)
	activeSeasonID := config.GetActiveSeasonID()

	switch grant.Type {
	case domain.GrantTypeAccess:
		return wrapRevokeError(queries.RevokeProfileAccess(ctx, profile.MinecraftUUID, grant.ID, revokedBy))
	case domain.GrantTypeNameColor:
		if err := queries.RevokeProfileNameColorOption(ctx, profile.ID, grant.ID, revokedBy); err != nil {
			return wrapRevokeError(err)
		}
		if profile.NameColorID != grant.ItemID {
			return nil
		}

		remaining, err := queries.FindProfileNameColorOptionsByProfileID(ctx, profile.ID, &activeSeasonID)
		if err != nil {
			return utils.NewInternalServerError("failed to find profile name color options", err)
		}
		if slices.ContainsFunc(remaining, func(option *domain.ProfileNameColorOption) bool { return option.NameColorID == grant.ItemID }) {
			return nil
		}
		return s.profileCosmeticsService.SelectProfileNameColor(ctx, queries, profile.ID, config.GetDefaultNameColorID())
	case domain.GrantTypeNamePrefix:
		if err := queries.RevokeProfileNamePrefixOption(ctx, profile.ID, grant.ID, revokedBy); err != nil {
			return wrapRevokeError(err)
		}

		selected, err := queries.FindProfilePrefixesByProfileID(ctx, profile.ID)
		if err != nil {
			return utils.NewInternalServerError("failed to find profile prefixes", err)
		}
		if !slices.ContainsFunc(selected, func(prefix *domain.ProfilePrefix) bool {
			return prefix.Type == grant.PrefixType && prefix.NamePrefixID == grant.ItemID
		}) {
			return nil
		}

		remaining, err := queries.FindProfileNamePrefixOptionsByProfileIDAndType(ctx, profile.ID, grant.PrefixType, &activeSeasonID)
		if err != nil {
			return utils.NewInternalServerError("failed to find profile name prefix options", err)
		}
		if slices.ContainsFunc(remaining, func(option *domain.ProfileNamePrefixOption) bool { return option.NamePrefixID == grant.ItemID }) {
			return nil
		}
		return s.profileCosmeticsService.ClearProfilePrefixByType(ctx, queries, profile.ID, grant.PrefixType)
	}
	return utils.NewBadRequestError("unknown grant type", nil)
}

func wrapRevokeError(err error) error {
	if err == nil {
		return nil
	}
	return utils.NewInternalServerError("failed to revoke grant", err)
}

// updateGame makes the Minecraft server match the database after a revoke.
func (s *adminGrantService) updateGame(ctx context.Context, profile *domain.Profile, grant *domain.Grant) error {
	if grant.Type != domain.GrantTypeAccess {
		return s.updatePrefix(ctx, profile.ID)
	}

	// Only the active season runs on the Minecraft server.
	activeSeasonID := config.GetActiveSeasonID()
	if grant.SeasonID == nil || *grant.SeasonID != activeSeasonID {
		return nil
	}

	hasAccess, err := s.accessService.CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, profile.MinecraftUUID, activeSeasonID)
	if err != nil || hasAccess {
		return err
	}
	return s.minecraftService.RemoveFromWhitelist(ctx, profile)
}

// updatePrefix sends the chat prefix built from the selected name color and prefixes to the Minecraft server.
func (s *adminGrantService) updatePrefix(ctx context.Context, profileID uuid.UUID) error {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return err
	}

	prefixes, err := s.profileCosmeticsService.GetProfilePrefixes(ctx, profileID)
	if err != nil {
		return err
	}

	var glythPrefix, specialPrefix *domain.NamePrefix
	for _, prefix := range prefixes {
		switch prefix.Type {
		case domain.ProfilePrefixTypeGlyth:
			glythPrefix = prefix.NamePrefix
		case domain.ProfilePrefixTypeSpecial:
			specialPrefix = prefix.NamePrefix
		}
	}

	formattedPrefix := s.profileCosmeticsService.GetProfileFullPrefix(ctx, profile.NameColor, glythPrefix, specialPrefix)
	return s.minecraftService.SetPrefixByMinecraftUUID(ctx, profile.MinecraftUUID, formattedPrefix)
}
