package clients

import (
	"context"

	"github.com/lania-smp/backend/internal/config"
	ory "github.com/ory/client-go"
)

type OryAPI interface {
	GetSession(ctx context.Context, cookies string) (*ory.Session, error)
	GetIdentity(ctx context.Context, identityID string) (*ory.Identity, error)
}

type oryAPI struct {
	client *ory.APIClient
}

func NewOryAPI(ctx context.Context) (OryAPI, error) {
	c := ory.NewConfiguration()
	c.Servers = ory.ServerConfigurations{
		{
			URL: config.GetOryUrl(),
		},
	}
	oryClient := ory.NewAPIClient(c)
	wrapper := &oryAPI{
		client: oryClient,
	}

	return wrapper, nil
}

func (api *oryAPI) GetSession(ctx context.Context, cookies string) (*ory.Session, error) {
	session, _, err := api.client.FrontendAPI.ToSession(ctx).Cookie(cookies).Execute()
	return session, err
}

func (api *oryAPI) GetIdentity(ctx context.Context, identityID string) (*ory.Identity, error) {
	identity, _, err := api.client.IdentityAPI.GetIdentity(ctx, identityID).Execute()
	return identity, err
}
