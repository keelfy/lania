package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

// FulfillmentService hands a product over to a profile. Orders and admins share it, so a product
// gives the same result whoever grants it.
type FulfillmentService interface {
	// GrantProduct gives the product to the profile for the season.
	// accessSource tells how a season access was obtained. orderItemID is nil when the product is not from an order.
	GrantProduct(ctx context.Context, queries sql.Queries, profileID, seasonID uuid.UUID, product *domain.Product, accessSource domain.AccessSource, orderItemID *uuid.UUID) error
}

type fulfillmentService struct {
	accessService           AccessService
	profileCosmeticsService ProfileCosmeticsService
	profileService          ProfileService
}

func NewFulfillmentService(
	accessService AccessService,
	profileCosmeticsService ProfileCosmeticsService,
	profileService ProfileService,
) FulfillmentService {
	return &fulfillmentService{
		accessService:           accessService,
		profileCosmeticsService: profileCosmeticsService,
		profileService:          profileService,
	}
}

func (s *fulfillmentService) GrantProduct(ctx context.Context, queries sql.Queries, profileID, seasonID uuid.UUID, product *domain.Product, accessSource domain.AccessSource, orderItemID *uuid.UUID) error {
	switch product.Category {
	case domain.ProductCategoryUpgrade:
		var metadata domain.UpgradeProductMetadata
		err := json.Unmarshal(product.Metadata, &metadata)
		if err != nil {
			return err
		}

		profile, err := s.profileService.GetProfileByID(ctx, profileID)
		if err != nil {
			return err
		}

		if metadata.Action == domain.ProductUpgradeActionSeasonAccess {
			// TODO: check if profile has access for the season
			return s.accessService.ObtainAccessForProfile(ctx, seasonID, profile, accessSource)
		} else {
			return fmt.Errorf("unknown upgrade action: %s", metadata.Action)
		}
	case domain.ProductCategoryNameColor:
		var metadata domain.NameColorProductMetadata
		err := json.Unmarshal(product.Metadata, &metadata)
		if err != nil {
			return err
		}
		return s.profileCosmeticsService.AddProfileNameColorOption(ctx, queries, profileID, metadata.NameColorID, &seasonID, orderItemID)
	case domain.ProductCategoryNamePrefix:
		var metadata domain.NamePrefixProductMetadata
		err := json.Unmarshal(product.Metadata, &metadata)
		if err != nil {
			return err
		}
		return s.profileCosmeticsService.AddProfileNameGlythOption(ctx, queries, profileID, metadata.NamePrefixID, &seasonID, orderItemID)
	default:
		return fmt.Errorf("unknown product category: %s", product.Category)
	}
}
