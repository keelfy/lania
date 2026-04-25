package presenter

import (
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	plandomain "github.com/lania-smp/backend/internal/domain/plan"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentProfileCosmetics(nameColor *domain.NameColor, glyth *domain.NamePrefix, special *domain.NamePrefix) *responses.ProfileCosmetics {
	nameColorResponse := &responses.NameColor{
		ID:     nameColor.ID,
		Name:   nameColor.Name,
		Colors: nameColor.Metadata.Colors,
	}
	var glythResponse *responses.NamePrefix
	if glyth != nil {
		glythResponse = &responses.NamePrefix{
			ID:     glyth.ID,
			Name:   glyth.Name,
			Prefix: glyth.Metadata.Prefix,
			Image:  glyth.Metadata.Image,
		}
	}
	var specialResponse *responses.NamePrefix
	if special != nil {
		specialResponse = &responses.NamePrefix{
			ID:     special.ID,
			Name:   special.Name,
			Prefix: special.Metadata.Prefix,
			Image:  special.Metadata.Image,
		}
	}
	return &responses.ProfileCosmetics{
		Name: &responses.ProfileNameCosmetics{
			Colors:        nameColorResponse,
			GlythPrefix:   glythResponse,
			SpecialPrefix: specialResponse,
		},
	}
}

func PresentProfile(
	profile *domain.Profile,
	mojangUUID *uuid.UUID,
	accessStatus domain.AccessStatus,
	cosmetics *responses.ProfileCosmetics,
) *responses.Profile {
	return &responses.Profile{
		ID:            profile.ID,
		MinecraftUUID: profile.MinecraftUUID,
		Username:      profile.MinecraftUsername,
		Cosmetics:     cosmetics,
		AccessStatus:  string(accessStatus),
		MojangUUID:    mojangUUID,
	}
}

func PresentPublicProfile(
	profile *domain.Profile,
	mojangUUID *uuid.UUID,
	cosmetics *responses.ProfileCosmetics,
	isOnline bool,
) *responses.PublicProfile {
	return &responses.PublicProfile{
		ID:            profile.ID,
		MinecraftUUID: profile.MinecraftUUID,
		Username:      profile.MinecraftUsername,
		Cosmetics:     cosmetics,
		Role:          string(profile.Role),
		IsOnline:      isOnline,
		MojangUUID:    mojangUUID,
	}
}

func PresentProfileDetails(
	profile *domain.Profile,
	mojangUUID *uuid.UUID,
	accessStatus domain.AccessStatus,
	playtime *plandomain.Playtime,
	isOnline bool,
	isModelSlim bool,
	cosmetics *responses.ProfileCosmetics,
) *responses.ProfileDetails {
	return &responses.ProfileDetails{
		ID:            profile.ID,
		MinecraftUUID: profile.MinecraftUUID,
		Username:      profile.MinecraftUsername,
		Cosmetics:     cosmetics,
		IsSlimModel:   isModelSlim,
		FirstSeenAt:   playtime.FirstSessionStart,
		LastSeenAt:    playtime.LastSessionEnd,
		Role:          string(profile.Role),
		AccessStatus:  string(accessStatus),
		Playtime:      playtime.TotalPlaytime,
		IsOnline:      isOnline,
		MojangUUID:    mojangUUID,
	}
}

func PresentProfileCosmeticOptions(nameColorOptions []*domain.ProfileNameColorOption, glythPrefixOptions []*domain.ProfileNamePrefixOption, specialPrefixOptions []*domain.ProfileNamePrefixOption) *responses.ProfileCosmeticOptions {
	nameColors := make([]*responses.ProfileNameColorOption, len(nameColorOptions))
	for i, profileNameColor := range nameColorOptions {
		nameColors[i] = &responses.ProfileNameColorOption{
			ID:          profileNameColor.ID,
			NameColorID: profileNameColor.NameColorID,
			Name:        profileNameColor.NameColor.Name,
			Colors:      profileNameColor.NameColor.Metadata.Colors,
			ProfileID:   profileNameColor.ProfileID,
			ForSeasonID: profileNameColor.ForSeasonID,
		}
	}
	glyths := make([]*responses.ProfileNamePrefixOption, len(glythPrefixOptions))
	for i, profileNamePrefix := range glythPrefixOptions {
		glyths[i] = &responses.ProfileNamePrefixOption{
			ID:           profileNamePrefix.ID,
			NamePrefixID: profileNamePrefix.NamePrefixID,
			Name:         profileNamePrefix.NamePrefix.Name,
			Prefix:       profileNamePrefix.NamePrefix.Metadata.Prefix,
			Image:        profileNamePrefix.NamePrefix.Metadata.Image,
			ProfileID:    profileNamePrefix.ProfileID,
			ForSeasonID:  profileNamePrefix.ForSeasonID,
		}
	}
	specials := make([]*responses.ProfileNamePrefixOption, len(specialPrefixOptions))
	for i, profileNamePrefix := range specialPrefixOptions {
		specials[i] = &responses.ProfileNamePrefixOption{
			ID:           profileNamePrefix.ID,
			NamePrefixID: profileNamePrefix.NamePrefixID,
			Name:         profileNamePrefix.NamePrefix.Name,
			Prefix:       profileNamePrefix.NamePrefix.Metadata.Prefix,
			Image:        profileNamePrefix.NamePrefix.Metadata.Image,
			ProfileID:    profileNamePrefix.ProfileID,
			ForSeasonID:  profileNamePrefix.ForSeasonID,
		}
	}
	return &responses.ProfileCosmeticOptions{
		Name: &responses.ProfileNameCosmeticOptions{
			Colors:          nameColors,
			GlythPrefixes:   glyths,
			SpecialPrefixes: specials,
		},
	}
}
