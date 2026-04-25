package domain

import (
	"time"

	"github.com/google/uuid"
)

type IntegrationService string

const (
	IntegrationServiceDonationAlert IntegrationService = "donation_alerts"
)

type OAuth2Integration struct {
	ID           uuid.UUID
	ServiceName  IntegrationService
	AccessToken  string
	RefreshToken string
	UpdatedAt    time.Time
}
