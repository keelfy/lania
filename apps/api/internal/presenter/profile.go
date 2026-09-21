package presenter

import (
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

// PresentProfileCosmetics shows what a profile wears. A nil cosmetics stands for a profile that could not be read.
func PresentProfileCosmetics(cosmetics *domain.ProfileCosmetics) *responses.ProfileCosmetics {
	if cosmetics == nil {
		cosmetics = &domain.ProfileCosmetics{}
	}
	nameColor, glyth, special := cosmetics.NameColor, cosmetics.Glyth, cosmetics.Special

	var nameColorResponse *responses.NameColor
	if nameColor != nil {
		nameColorResponse = &responses.NameColor{
			ID:     nameColor.ID,
			Name:   nameColor.Name,
			Colors: nameColor.Metadata.Colors,
		}
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

func presentSeasonAccesses(statuses []*domain.SeasonAccessStatus) []*responses.SeasonAccess {
	res := make([]*responses.SeasonAccess, len(statuses))
	for i, status := range statuses {
		res[i] = &responses.SeasonAccess{SeasonID: status.SeasonID, Status: string(status.Status)}
	}
	return res
}

func PresentProfile(
	profile *domain.Profile,
	mojangUUID *uuid.UUID,
	accessStatus domain.AccessStatus,
	seasonAccesses []*domain.SeasonAccessStatus,
	cosmetics *responses.ProfileCosmetics,
) *responses.Profile {
	return &responses.Profile{
		ID:            profile.ID,
		MinecraftUUID: profile.MinecraftUUID,
		Username:      profile.MinecraftUsername,
		Cosmetics:     cosmetics,
		AccessStatus:  string(accessStatus),
		Accesses:      presentSeasonAccesses(seasonAccesses),
		MojangUUID:    mojangUUID,
	}
}

func PresentPublicProfile(
	profile *domain.Profile,
	mojangUUID *uuid.UUID,
	cosmetics *responses.ProfileCosmetics,
	isOnline bool,
	// playtime is summed over all seasons, in milliseconds.
	playtime int64,
	// lastSeenAt is the date in the season the list is for, nil when it is unknown.
	lastSeenAt *time.Time,
) *responses.PublicProfile {
	return &responses.PublicProfile{
		ID:            profile.ID,
		MinecraftUUID: profile.MinecraftUUID,
		Username:      profile.MinecraftUsername,
		Cosmetics:     cosmetics,
		Role:          string(profile.Role),
		IsOnline:      isOnline,
		Playtime:      playtime,
		LastSeenAt:    timeToMillis(lastSeenAt),
		MojangUUID:    mojangUUID,
	}
}

func timeToMillis(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	millis := t.UnixMilli()
	return &millis
}

func PresentProfileDetails(
	profile *domain.Profile,
	mojangUUID *uuid.UUID,
	accessStatus domain.AccessStatus,
	seasonAccesses []*domain.SeasonAccessStatus,
	// playtime is summed over all seasons, the current one is synced from the Minecraft server in background.
	playtime int64,
	isOnline bool,
	isModelSlim bool,
	cosmetics *responses.ProfileCosmetics,
	// lastSeenAt is the date in the season the profile is shown for, nil when it is unknown.
	lastSeenAt *time.Time,
) *responses.ProfileDetails {
	return &responses.ProfileDetails{
		ID:            profile.ID,
		MinecraftUUID: profile.MinecraftUUID,
		Username:      profile.MinecraftUsername,
		Cosmetics:     cosmetics,
		IsSlimModel:   isModelSlim,
		FirstSeenAt:   timeToMillis(profile.FirstSeenAt),
		LastSeenAt:    timeToMillis(lastSeenAt),
		Role:          string(profile.Role),
		AccessStatus:  string(accessStatus),
		Accesses:      presentSeasonAccesses(seasonAccesses),
		Playtime:      playtime,
		IsOnline:      isOnline,
		MojangUUID:    mojangUUID,
	}
}

// PresentProfileStats keeps the order of the given seasons.
func PresentProfileStats(stats []*domain.ProfileSeasonStats) *responses.ProfileStats {
	res := &responses.ProfileStats{Seasons: make([]*responses.ProfileSeasonStats, len(stats))}
	for i, seasonStats := range stats {
		res.TotalPlaytime += seasonStats.Playtime
		res.Seasons[i] = &responses.ProfileSeasonStats{
			SeasonID:   seasonStats.SeasonID,
			SeasonName: seasonStats.SeasonName,
			StartDate:  seasonStats.StartDate.UnixMilli(),
			EndDate:    timeToMillis(seasonStats.EndDate),
			IsActive:   seasonStats.IsActive,
			IsPrimary:  seasonStats.IsPrimary,
			Playtime:   seasonStats.Playtime,
		}
	}
	return res
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

func PresentProfileResync(resync *domain.ProfileResync) *responses.ProfileResync {
	seasons := make([]*responses.SeasonResync, len(resync.Seasons))
	for i, season := range resync.Seasons {
		parts := make([]*responses.PartResync, len(season.Parts))
		for j, part := range season.Parts {
			parts[j] = &responses.PartResync{Part: string(part.Part), OK: part.Err == nil}
			if part.Err != nil {
				message := utils.ExtractErrorMessage(part.Err)
				parts[j].Error = &message
			}
		}
		seasons[i] = &responses.SeasonResync{
			SeasonID: season.SeasonID, SeasonName: season.SeasonName, OK: season.OK(), Parts: parts,
		}
	}
	return &responses.ProfileResync{OK: resync.OK(), Seasons: seasons}
}
