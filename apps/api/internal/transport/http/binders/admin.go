package binders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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

var SeasonIDVariable = "seasonId"

var CosmeticIDVariable = "cosmeticId"
var ProductIDVariable = "productId"

func optionalTrimmed(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func parseSeasonDate(value string) (time.Time, error) {
	return time.Parse(time.DateOnly, value)
}

func BindSaveSeason(r *http.Request) (*commands.SaveSeasonCommand, error) {
	req := &requests.SaveSeason{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}

	startDate, err := parseSeasonDate(req.StartDate)
	if err != nil {
		return nil, utils.NewBadRequestError("startDate must use YYYY-MM-DD", err)
	}

	var endDate *time.Time
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		parsed, err := parseSeasonDate(*req.EndDate)
		if err != nil {
			return nil, utils.NewBadRequestError("endDate must use YYYY-MM-DD", err)
		}
		endDate = &parsed
	}

	return &commands.SaveSeasonCommand{
		Name:             strings.TrimSpace(req.Name),
		PreviewImage:     optionalTrimmed(req.PreviewImage),
		StartDate:        startDate,
		EndDate:          endDate,
		PublicAddress:    optionalTrimmed(req.PublicAddress),
		ShellAddress:     optionalTrimmed(req.ShellAddress),
		PlanURL:          optionalTrimmed(req.PlanURL),
		IsActive:         req.IsActive,
		IsPrimary:        req.IsPrimary,
		Preregistration:  req.Preregistration,
		FreeRegistration: req.FreeRegistration,
	}, nil
}

func BindUpdateSeason(r *http.Request) (*commands.SaveSeasonCommand, error) {
	cmd, err := BindSaveSeason(r)
	if err != nil {
		return nil, err
	}
	cmd.ID, err = BindPathVariableAsUUID(r, SeasonIDVariable)
	return cmd, err
}

func BindSaveNameColor(r *http.Request) (*commands.SaveNameColorCommand, error) {
	var req requests.SaveNameColor
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}
	colors := make([]string, 0, len(req.Colors))
	for _, color := range req.Colors {
		colors = append(colors, strings.TrimSpace(color))
	}
	return &commands.SaveNameColorCommand{Name: strings.TrimSpace(req.Name), Colors: colors}, nil
}

func BindUpdateNameColor(r *http.Request) (*commands.SaveNameColorCommand, error) {
	cmd, err := BindSaveNameColor(r)
	if err != nil {
		return nil, err
	}
	cmd.ID, err = BindPathVariableAsUUID(r, CosmeticIDVariable)
	return cmd, err
}

func BindSaveNamePrefix(r *http.Request) (*commands.SaveNamePrefixCommand, error) {
	var req requests.SaveNamePrefix
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}
	return &commands.SaveNamePrefixCommand{Name: strings.TrimSpace(req.Name), Prefix: strings.TrimSpace(req.Prefix), Image: strings.TrimSpace(req.Image), NoSpace: req.NoSpace}, nil
}

func BindUpdateNamePrefix(r *http.Request) (*commands.SaveNamePrefixCommand, error) {
	cmd, err := BindSaveNamePrefix(r)
	if err != nil {
		return nil, err
	}
	cmd.ID, err = BindPathVariableAsUUID(r, CosmeticIDVariable)
	return cmd, err
}

func BindSaveProduct(r *http.Request) (*commands.SaveProductCommand, error) {
	var req requests.SaveProduct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}
	localizations := make([]commands.SaveProductLocalization, len(req.Localizations))
	for i, localization := range req.Localizations {
		localizations[i] = commands.SaveProductLocalization{Locale: strings.ToLower(strings.TrimSpace(localization.Locale)), Name: strings.TrimSpace(localization.Name), Description: strings.TrimSpace(localization.Description)}
	}
	return &commands.SaveProductCommand{Category: domain.ProductCategory(strings.TrimSpace(req.Category)), CosmeticID: req.CosmeticID, PriceName: domain.ProductPriceName(strings.TrimSpace(req.PriceName)), IsActive: req.IsActive, EasyDonateProductID: req.EasyDonateProductID, Localizations: localizations}, nil
}

func BindUpdateProduct(r *http.Request) (*commands.SaveProductCommand, error) {
	cmd, err := BindSaveProduct(r)
	if err != nil {
		return nil, err
	}
	cmd.ID, err = BindPathVariableAsUUID(r, ProductIDVariable)
	return cmd, err
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

func BindSetProfileRole(r *http.Request) (*commands.SetProfileRoleCommand, error) {
	profileID, err := BindPathVariableAsUUID(r, ProfileIDVariable)
	if err != nil {
		return nil, err
	}

	req := &requests.SetProfileRole{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}

	return &commands.SetProfileRoleCommand{
		ProfileID: profileID,
		Role:      domain.Role(strings.TrimSpace(req.Role)),
	}, nil
}

var TargetProfileIDQueryParam = "targetProfileId"

// BindPreviewMergeProfiles reads the source profile from the path and the target from the query,
// for the preview GET that has no body.
func BindPreviewMergeProfiles(r *http.Request) (*commands.MergeProfilesCommand, error) {
	sourceProfileID, err := BindPathVariableAsUUID(r, ProfileIDVariable)
	if err != nil {
		return nil, err
	}

	targetProfileIDStr, err := BindMandatoryQueryParamAsString(r, TargetProfileIDQueryParam)
	if err != nil {
		return nil, err
	}
	targetProfileID, err := uuid.Parse(targetProfileIDStr)
	if err != nil {
		return nil, utils.NewBadRequestError(fmt.Sprintf("query parameter %v is not a valid UUID", TargetProfileIDQueryParam), err)
	}

	return &commands.MergeProfilesCommand{SourceProfileID: sourceProfileID, TargetProfileID: targetProfileID}, nil
}

func BindMergeProfiles(r *http.Request) (*commands.MergeProfilesCommand, error) {
	sourceProfileID, err := BindPathVariableAsUUID(r, ProfileIDVariable)
	if err != nil {
		return nil, err
	}

	req := &requests.MergeProfiles{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}

	return &commands.MergeProfilesCommand{SourceProfileID: sourceProfileID, TargetProfileID: req.TargetProfileID}, nil
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
