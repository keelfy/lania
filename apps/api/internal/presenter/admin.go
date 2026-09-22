package presenter

import (
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentAdminUser(user *domain.User) *responses.AdminUser {
	return &responses.AdminUser{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		AvatarURL: user.AvatarURL,
		Role:      string(user.Role),
		CreatedAt: timeToMillis(user.CreatedAt),
	}
}

func PresentAdminUserDetails(user *domain.User) *responses.AdminUserDetails {
	profiles := make([]*responses.AdminProfile, len(user.Profiles))
	for i, profile := range user.Profiles {
		profiles[i] = PresentAdminProfile(profile)
	}
	return &responses.AdminUserDetails{
		AdminUser: *PresentAdminUser(user),
		Profiles:  profiles,
	}
}

func PresentAdminProfile(profile *domain.Profile) *responses.AdminProfile {
	return &responses.AdminProfile{
		ID:            profile.ID,
		MinecraftUUID: profile.MinecraftUUID,
		Username:      profile.MinecraftUsername,
		OwnerUserID:   profile.OwnerUserID,
		Role:          string(profile.Role),
		CreatedAt:     profile.CreatedAt.UnixMilli(),
	}
}

func PresentAdminProfileDetails(profile *domain.Profile, owner *domain.User) *responses.AdminProfileDetails {
	details := &responses.AdminProfileDetails{AdminProfile: *PresentAdminProfile(profile)}
	if owner != nil {
		details.Owner = PresentAdminUser(owner)
	}
	return details
}

func presentProfileMergeBlockers(blockers []*domain.ProfileMergeBlocker) []*responses.ProfileMergeBlocker {
	if len(blockers) == 0 {
		return nil
	}
	res := make([]*responses.ProfileMergeBlocker, len(blockers))
	for i, blocker := range blockers {
		res[i] = &responses.ProfileMergeBlocker{Kind: string(blocker.Kind), SeasonNames: blocker.SeasonNames}
	}
	return res
}

// PresentProfileMergeSummary renders a preview or the outcome of a merge. resync is nil for a preview.
func PresentProfileMergeSummary(summary *domain.ProfileMergeSummary, resync *domain.ProfileResync) *responses.ProfileMergeSummary {
	res := &responses.ProfileMergeSummary{
		SourceProfileID: summary.SourceProfileID,
		SourceUsername:  summary.SourceUsername,
		TargetProfileID: summary.TargetProfileID,
		TargetUsername:  summary.TargetUsername,
		RoleBefore:      string(summary.RoleBefore),
		RoleAfter:       string(summary.RoleAfter),
		OwnerUserID:     summary.OwnerUserID,
		Counts: responses.ProfileMergeCounts{
			PlaytimeMoved:            summary.Counts.PlaytimeMoved,
			PlaytimeSummed:           summary.Counts.PlaytimeSummed,
			AccessesMoved:            summary.Counts.AccessesMoved,
			AccessesDropped:          summary.Counts.AccessesDropped,
			ViolationsMoved:          summary.Counts.ViolationsMoved,
			NameColorOptionsMoved:    summary.Counts.NameColorOptionsMoved,
			NameColorOptionsDropped:  summary.Counts.NameColorOptionsDropped,
			NamePrefixOptionsMoved:   summary.Counts.NamePrefixOptionsMoved,
			NamePrefixOptionsDropped: summary.Counts.NamePrefixOptionsDropped,
			SeasonCosmeticsMoved:     summary.Counts.SeasonCosmeticsMoved,
			SeasonCosmeticsDropped:   summary.Counts.SeasonCosmeticsDropped,
			PrefixesMoved:            summary.Counts.PrefixesMoved,
			PrefixesDropped:          summary.Counts.PrefixesDropped,
			OrderItemsMoved:          summary.Counts.OrderItemsMoved,
			BasketItemsMoved:         summary.Counts.BasketItemsMoved,
			BasketItemsDropped:       summary.Counts.BasketItemsDropped,
			NotificationsRepointed:   summary.Counts.NotificationsRepointed,
		},
		Blockers: presentProfileMergeBlockers(summary.Blockers),
		CanMerge: summary.CanMerge(),
	}
	if resync != nil {
		res.Resync = PresentProfileResync(resync)
	}
	return res
}

func PresentProfileMerges(merges []*domain.ProfileMerge) []*responses.ProfileMerge {
	res := make([]*responses.ProfileMerge, len(merges))
	for i, merge := range merges {
		res[i] = &responses.ProfileMerge{
			ID:                  merge.ID,
			SourceProfileID:     merge.SourceProfileID,
			SourceMinecraftUUID: merge.SourceMinecraftUUID,
			SourceUsername:      merge.SourceUsername,
			MergedBy:            merge.MergedBy,
			CreatedAt:           merge.CreatedAt.UnixMilli(),
		}
	}
	return res
}

func PresentAdminGrants(grants []*domain.Grant) []*responses.AdminGrant {
	res := make([]*responses.AdminGrant, len(grants))
	for i, grant := range grants {
		res[i] = &responses.AdminGrant{
			ID:          grant.ID,
			Type:        string(grant.Type),
			SeasonID:    grant.SeasonID,
			Name:        grant.Name,
			PrefixType:  string(grant.PrefixType),
			Source:      string(grant.Source),
			OrderItemID: grant.OrderItemID,
			GrantedBy:   grant.GrantedBy,
			CreatedAt:   grant.CreatedAt.UnixMilli(),
			RevokedAt:   timeToMillis(grant.RevokedAt),
			RevokedBy:   grant.RevokedBy,
		}
	}
	return res
}

func PresentAdminCosmeticsCatalog(catalog *domain.CosmeticsCatalog) *responses.AdminCosmeticsCatalog {
	res := &responses.AdminCosmeticsCatalog{
		NameColors:   make([]*responses.AdminNameColor, len(catalog.NameColors)),
		NamePrefixes: make([]*responses.AdminNamePrefix, len(catalog.NamePrefixes)),
	}
	for i, nameColor := range catalog.NameColors {
		colors := nameColor.Metadata.Colors
		if colors == nil {
			colors = []string{}
		}
		res.NameColors[i] = &responses.AdminNameColor{ID: nameColor.ID, Name: nameColor.Name, Colors: colors}
	}
	for i, namePrefix := range catalog.NamePrefixes {
		res.NamePrefixes[i] = &responses.AdminNamePrefix{ID: namePrefix.ID, Name: namePrefix.Name, Image: namePrefix.Metadata.Image, Prefix: namePrefix.Metadata.Prefix, NoSpace: namePrefix.Metadata.NoSpace}
	}
	return res
}

func PresentAdminProducts(products []*domain.Product) []*responses.AdminProduct {
	res := make([]*responses.AdminProduct, len(products))
	for i, product := range products {
		localizations := make([]*responses.AdminProductLocalization, len(product.Localizations))
		for j, localization := range product.Localizations {
			localizations[j] = &responses.AdminProductLocalization{Locale: localization.Locale, Name: localization.Name, Description: localization.Description}
		}
		prices := make([]*responses.AdminProductPrice, len(product.Prices))
		for j, price := range product.Prices {
			prices[j] = &responses.AdminProductPrice{Currency: string(price.Currency), Amount: price.Amount}
		}
		res[i] = &responses.AdminProduct{ID: product.ID, Category: string(product.Category), PriceName: string(product.PriceName), Metadata: product.Metadata, IsActive: product.IsActive, EasyDonateProductID: product.EasyDonateProductID, SoldCount: product.SoldCount, Localizations: localizations, Prices: prices}
	}
	return res
}
