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
		res.NamePrefixes[i] = &responses.AdminNamePrefix{ID: namePrefix.ID, Name: namePrefix.Name, Image: namePrefix.Metadata.Image}
	}
	return res
}
