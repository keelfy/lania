package sql

import (
	"context"

	"github.com/lania-smp/backend/internal/domain"
)

const findOAuth2IntegrationByServiceName = `
SELECT id, service_name, access_token, refresh_token, updated_at 
FROM oauth2_integrations 
WHERE service_name = ?
`

func (q *queries) FindOAuth2IntegrationByServiceName(ctx context.Context, serviceName domain.IntegrationService) (*domain.OAuth2Integration, error) {
	row := q.x.QueryRowContext(ctx, findOAuth2IntegrationByServiceName, serviceName)
	var integration domain.OAuth2Integration
	err := row.Scan(
		&integration.ID,
		&integration.ServiceName,
		&integration.AccessToken,
		&integration.RefreshToken,
		&integration.UpdatedAt,
	)
	return &integration, err
}

const updateOAuth2Integration = `
UPDATE oauth2_integrations 
SET access_token = ?, refresh_token = ?, updated_at = now() 
WHERE service_name = ?
`

type UpdateOAuth2IntegrationParams struct {
	AccessToken  string
	RefreshToken string
	ServiceName  domain.IntegrationService
}

func (q *queries) UpdateOAuth2Integration(ctx context.Context, arg UpdateOAuth2IntegrationParams) error {
	_, err := q.x.ExecContext(ctx, updateOAuth2Integration, arg.AccessToken, arg.RefreshToken, arg.ServiceName)
	return err
}
