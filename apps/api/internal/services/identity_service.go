package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/utils"
	ory "github.com/ory/client-go"
)

type IdentityService interface {
	GetIdentityByID(ctx context.Context, id uuid.UUID) (*ory.Identity, error)
}

type identityService struct {
	oryClient clients.OryAPI
}

func NewIdentityService(
	oryClient clients.OryAPI,
) IdentityService {
	return &identityService{
		oryClient: oryClient,
	}
}

func (s *identityService) GetIdentityByID(ctx context.Context, id uuid.UUID) (*ory.Identity, error) {
	identity, err := s.oryClient.GetIdentity(ctx, id.String())
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get identity", err)
	}
	return identity, nil
}
