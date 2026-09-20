package clients

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/lania-smp/backend/internal/config"
	ory "github.com/ory/client-go"
)

// ErrIdentityNotFound is returned when Kratos has no identity with the requested ID.
var ErrIdentityNotFound = errors.New("identity not found")

type OryAPI interface {
	GetSession(ctx context.Context, cookies string) (*ory.Session, error)
	// GetIdentity returns ErrIdentityNotFound when there is no such identity.
	GetIdentity(ctx context.Context, identityID string) (*ory.Identity, error)
	// ListIdentities returns a page of identities and the token of the next page, empty on the last page.
	// credentialsIdentifier narrows the list down to the identities that log in with exactly this identifier.
	ListIdentities(ctx context.Context, pageSize int64, pageToken, credentialsIdentifier string) ([]ory.Identity, string, error)
	// PatchIdentityMetadataPublic merges the values into metadata_public of the identity and keeps other keys.
	PatchIdentityMetadataPublic(ctx context.Context, identityID string, values map[string]any) error
}

type oryAPI struct {
	// client talks to the public Kratos API on behalf of a browser session.
	client *ory.APIClient
	// adminClient talks to the Kratos admin API. Only the API itself may hold this address.
	adminClient *ory.APIClient
}

func NewOryAPI(ctx context.Context) (OryAPI, error) {
	wrapper := &oryAPI{
		client:      newOryClient(config.GetOryUrl()),
		adminClient: newOryClient(config.GetOryAdminUrl()),
	}

	return wrapper, nil
}

func newOryClient(serverURL string) *ory.APIClient {
	c := ory.NewConfiguration()
	c.Servers = ory.ServerConfigurations{
		{
			URL: serverURL,
		},
	}
	return ory.NewAPIClient(c)
}

func (api *oryAPI) GetSession(ctx context.Context, cookies string) (*ory.Session, error) {
	session, _, err := api.client.FrontendAPI.ToSession(ctx).Cookie(cookies).Execute()
	return session, err
}

func (api *oryAPI) GetIdentity(ctx context.Context, identityID string) (*ory.Identity, error) {
	identity, res, err := api.adminClient.IdentityAPI.GetIdentity(ctx, identityID).Execute()
	if err != nil && res != nil && res.StatusCode == http.StatusNotFound {
		return nil, ErrIdentityNotFound
	}
	return identity, err
}

func (api *oryAPI) ListIdentities(ctx context.Context, pageSize int64, pageToken, credentialsIdentifier string) ([]ory.Identity, string, error) {
	req := api.adminClient.IdentityAPI.ListIdentities(ctx).PageSize(pageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	if credentialsIdentifier != "" {
		req = req.CredentialsIdentifier(credentialsIdentifier)
	}

	identities, res, err := req.Execute()
	if err != nil {
		return nil, "", err
	}
	return identities, nextPageToken(res.Header), nil
}

func (api *oryAPI) PatchIdentityMetadataPublic(ctx context.Context, identityID string, values map[string]any) error {
	identity, err := api.GetIdentity(ctx, identityID)
	if err != nil {
		return err
	}

	// A JSON patch cannot add a key to a null metadata_public, so the whole object is replaced.
	merged := make(map[string]any, len(identity.MetadataPublic)+len(values))
	for key, value := range identity.MetadataPublic {
		merged[key] = value
	}
	for key, value := range values {
		merged[key] = value
	}

	_, _, err = api.adminClient.IdentityAPI.PatchIdentity(ctx, identityID).
		JsonPatch([]ory.JsonPatch{{Op: "add", Path: "/metadata_public", Value: merged}}).
		Execute()
	return err
}

// nextPageToken reads the token of the next page from the Link header of a Kratos list response:
// <https://kratos/admin/identities?page_size=250&page_token=abc>; rel="next"
func nextPageToken(header http.Header) string {
	for _, value := range header.Values("Link") {
		for _, link := range strings.Split(value, ",") {
			target, params, found := strings.Cut(strings.TrimSpace(link), ";")
			if !found || !strings.Contains(params, `rel="next"`) {
				continue
			}
			linkURL, err := url.Parse(strings.Trim(strings.TrimSpace(target), "<>"))
			if err != nil {
				return ""
			}
			return linkURL.Query().Get("page_token")
		}
	}
	return ""
}
