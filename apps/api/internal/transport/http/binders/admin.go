package binders

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

var PageTokenQueryParam = "pageToken"

// MaxEmailLength is the longest valid email address.
const MaxEmailLength = 254

// BindEmailSearch returns the trimmed search query, cut to MaxEmailLength characters.
func BindEmailSearch(r *http.Request) string {
	search := []rune(strings.TrimSpace(r.URL.Query().Get(SearchQueryParam)))
	if len(search) > MaxEmailLength {
		search = search[:MaxEmailLength]
	}
	return string(search)
}

// BindPageToken returns the opaque token of the requested page, empty for the first page.
func BindPageToken(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get(PageTokenQueryParam))
}

var ProfileIDVariable = "profileId"

func BindTransferProfileOwner(r *http.Request) (*commands.TransferProfileOwnerCommand, error) {
	profileID, err := BindPathVariableAsUUID(r, ProfileIDVariable)
	if err != nil {
		return nil, err
	}

	req := &requests.TransferProfileOwner{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}

	return &commands.TransferProfileOwnerCommand{
		ProfileID: profileID,
		Email:     strings.TrimSpace(req.Email),
	}, nil
}

var (
	GrantTypeVariable = "grantType"
	GrantIDVariable   = "grantId"
)

func BindGrantProduct(r *http.Request) (*commands.GrantProductCommand, error) {
	profileID, err := BindPathVariableAsUUID(r, ProfileIDVariable)
	if err != nil {
		return nil, err
	}

	req := &requests.GrantProduct{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}

	return &commands.GrantProductCommand{
		ProfileID: profileID,
		ProductID: req.ProductID,
		SeasonID:  req.SeasonID,
	}, nil
}

func BindRevokeGrant(r *http.Request) (*commands.RevokeGrantCommand, error) {
	profileID, err := BindPathVariableAsUUID(r, ProfileIDVariable)
	if err != nil {
		return nil, err
	}
	grantType, err := BindPathVariable(r, GrantTypeVariable)
	if err != nil {
		return nil, err
	}
	grantID, err := BindPathVariableAsUUID(r, GrantIDVariable)
	if err != nil {
		return nil, err
	}

	return &commands.RevokeGrantCommand{
		ProfileID: profileID,
		GrantType: domain.GrantType(grantType),
		GrantID:   grantID,
	}, nil
}

func BindGrantCosmetic(r *http.Request) (*commands.GrantCosmeticCommand, error) {
	profileID, err := BindPathVariableAsUUID(r, ProfileIDVariable)
	if err != nil {
		return nil, err
	}

	req := &requests.GrantCosmetic{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}

	return &commands.GrantCosmeticCommand{
		ProfileID:  profileID,
		Type:       domain.GrantType(req.Type),
		ItemID:     req.ItemID,
		PrefixType: domain.ProfilePrefixType(req.PrefixType),
		SeasonID:   req.SeasonID,
	}, nil
}
