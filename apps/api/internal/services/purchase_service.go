package services

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

type PurchaseService interface {
	GetPurchasesByProducts(ctx context.Context, userID uuid.UUID, seasonID uuid.UUID, products []*domain.Product) ([]*domain.PurchasedProduct, error)
}

type purchaseService struct {
	storage                 storage.MainStorage
	accessService           AccessService
	profileCosmeticsService ProfileCosmeticsService
	profilePrivilegeService ProfilePrivilegeService
}

func NewPurchaseService(
	storage storage.MainStorage,
	accessService AccessService,
	profileCosmeticsService ProfileCosmeticsService,
	profilePrivilegeService ProfilePrivilegeService,
) PurchaseService {
	return &purchaseService{
		storage:                 storage,
		accessService:           accessService,
		profileCosmeticsService: profileCosmeticsService,
		profilePrivilegeService: profilePrivilegeService,
	}
}

func (s *purchaseService) GetPurchasesByProducts(ctx context.Context, userID uuid.UUID, seasonID uuid.UUID, products []*domain.Product) ([]*domain.PurchasedProduct, error) {
	hasUpgradeProduct := false
	hasNameColorProduct := false
	hasNamePrefixProduct := false
	hasPrivilegeProduct := false

	resList := make([]*domain.PurchasedProduct, 0)
	for _, product := range products {
		switch product.Category {
		case domain.ProductCategoryUpgrade:
			hasUpgradeProduct = true
		case domain.ProductCategoryNameColor:
			hasNameColorProduct = true
		case domain.ProductCategoryNamePrefix:
			hasNamePrefixProduct = true
		case domain.ProductCategoryPrivilege:
			hasPrivilegeProduct = true
		}
	}

	if hasUpgradeProduct {
		profileIDs, err := s.accessService.GetProfileIDsWithAccessBySeasonIDAndOwnerUserID(ctx, seasonID, userID)
		if err != nil {
			return nil, utils.NewInternalServerError("failed to get profile ids with access by season id and owner user id", err)
		}

		for _, profileID := range profileIDs {
			for _, product := range products {
				switch product.Category {
				case domain.ProductCategoryUpgrade:
					var metadata domain.UpgradeProductMetadata
					err := json.Unmarshal(product.Metadata, &metadata)
					if err != nil {
						return nil, utils.NewInternalServerError("failed to unmarshal upgrade product metadata", err)
					}

					if metadata.Action == domain.ProductUpgradeActionSeasonAccess {
						resList = append(resList, &domain.PurchasedProduct{
							ProductID: product.ID,
							ProfileID: profileID,
							SeasonID:  seasonID,
						})
					}
				}
			}
		}
	}

	if hasNameColorProduct {
		cosmetics, err := s.profileCosmeticsService.GetProfileNameColorOptionsByProfileOwnerUserID(ctx, userID, &seasonID)
		if err != nil {
			return nil, utils.NewInternalServerError("failed to get profile name color options by profile owner user id", err)
		}

		for _, cosmetic := range cosmetics {
			for _, product := range products {
				switch product.Category {
				case domain.ProductCategoryNameColor:
					var metadata domain.NameColorProductMetadata
					err := json.Unmarshal(product.Metadata, &metadata)
					if err != nil {
						return nil, utils.NewInternalServerError("failed to unmarshal name color product metadata", err)
					}

					if metadata.NameColorID == cosmetic.NameColorID {
						resList = append(resList, &domain.PurchasedProduct{
							ProductID: product.ID,
							ProfileID: cosmetic.ProfileID,
							SeasonID:  seasonID,
						})
					}
				}
			}
		}
	}

	if hasNamePrefixProduct {
		cosmetics, err := s.profileCosmeticsService.GetProfileNamePrefixOptionsByProfileOwnerUserIDAndType(ctx, userID, domain.ProfilePrefixTypeGlyth, &seasonID)
		if err != nil {
			return nil, utils.NewInternalServerError("failed to get profile name prefix options by profile owner user id and type", err)
		}

		for _, cosmetic := range cosmetics {
			for _, product := range products {
				switch product.Category {
				case domain.ProductCategoryNamePrefix:
					var metadata domain.NamePrefixProductMetadata
					err := json.Unmarshal(product.Metadata, &metadata)
					if err != nil {
						return nil, utils.NewInternalServerError("failed to unmarshal name prefix product metadata", err)
					}

					if metadata.NamePrefixID == cosmetic.NamePrefixID {
						resList = append(resList, &domain.PurchasedProduct{
							ProductID: product.ID,
							ProfileID: cosmetic.ProfileID,
							SeasonID:  seasonID,
						})
					}
				}
			}
		}
	}

	if hasPrivilegeProduct {
		privileges, err := s.profilePrivilegeService.GetProfilePrivilegesByOwnerUserID(ctx, userID, seasonID)
		if err != nil {
			return nil, utils.NewInternalServerError("failed to get profile privileges by profile owner user id", err)
		}

		for _, privilege := range privileges {
			for _, product := range products {
				if product.Category != domain.ProductCategoryPrivilege {
					continue
				}
				var metadata domain.PrivilegeProductMetadata
				if err := json.Unmarshal(product.Metadata, &metadata); err != nil {
					return nil, utils.NewInternalServerError("failed to unmarshal privilege product metadata", err)
				}

				if metadata.PrivilegeID == privilege.PrivilegeID {
					resList = append(resList, &domain.PurchasedProduct{
						ProductID: product.ID,
						ProfileID: privilege.ProfileID,
						SeasonID:  seasonID,
					})
				}
			}
		}
	}

	return resList, nil
}
