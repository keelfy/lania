package services

import (
	"context"
	"fmt"

	"github.com/lania-smp/backend/internal/clients"
)

type KratosService interface {
	GetOidcProviderConfig(ctx context.Context, userID string, providerID string) (*OidcProviderConfig, error)
}

type kratosService struct {
	oryAPI clients.OryAPI
}

func NewKratosService(oryAPI clients.OryAPI) KratosService {
	return &kratosService{
		oryAPI: oryAPI,
	}
}

type OidcProviderConfig struct {
	Provider            string `json:"provider,omitempty"`
	InitialIDToken      string `json:"initial_id_token,omitempty"`
	InitialAccessToken  string `json:"initial_access_token,omitempty"`
	InitialRefreshToken string `json:"initial_refresh_token,omitempty"`
	Subject             string `json:"subject,omitempty"`
	Organization        string `json:"organization,omitempty"`
}

func (s *kratosService) GetOidcProviderConfig(ctx context.Context, userID string, providerID string) (*OidcProviderConfig, error) {
	identity, err := s.oryAPI.GetIdentity(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting identity: %w", err)
	}

	var config *OidcProviderConfig
	credentials := identity.GetCredentials()

	for _, credential := range credentials {
		if credential.HasConfig() {
			providers := credential.GetConfig()["providers"].([]OidcProviderConfig)
			for _, provider := range providers {
				if provider.Provider == providerID {
					config = &provider
					break
				}
			}
		}
	}
	return config, nil
}
