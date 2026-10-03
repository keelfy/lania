package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func TestPurchaseService_PrivilegesAreOwnedPerPrivilege(t *testing.T) {
	ctx := context.Background()
	userID, seasonID, profileID := uuid.New(), uuid.New(), uuid.New()
	homes, fly := uuid.New(), uuid.New()

	privilegeProduct := func(privilegeID uuid.UUID) *domain.Product {
		metadata, _ := json.Marshal(domain.PrivilegeProductMetadata{PrivilegeID: privilegeID})
		return &domain.Product{ID: uuid.New(), Category: domain.ProductCategoryPrivilege, Metadata: metadata}
	}
	homesProduct, flyProduct := privilegeProduct(homes), privilegeProduct(fly)

	privileges := &recordingPrivileges{owned: []*domain.ProfilePrivilege{{ProfileID: profileID, PrivilegeID: homes, SeasonID: seasonID}}}
	service := NewPurchaseService(nil, nil, nil, privileges)

	purchased, err := service.GetPurchasesByProducts(ctx, userID, seasonID, []*domain.Product{homesProduct, flyProduct})
	if err != nil {
		t.Fatal(err)
	}

	if len(purchased) != 1 || purchased[0].ProductID != homesProduct.ID || purchased[0].ProfileID != profileID || purchased[0].SeasonID != seasonID {
		t.Fatalf("purchased = %+v, want only the product of the owned privilege for its profile and season", purchased)
	}
}
