package services

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type ProfileCosmeticsService interface {
	GetProfileNameColorOptions(ctx context.Context, profileID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error)
	GetProfileNamePrefixOptionsByProfileIDAndType(ctx context.Context, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error)
	GetProfileNameColorOptionByIDAndProfileID(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, seasonID *uuid.UUID) (*domain.ProfileNameColorOption, error)
	GetProfileNamePrefixOptionByIDAndProfileIDAndType(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) (*domain.ProfileNamePrefixOption, error)
	// SelectProfileNameColor, SelectProfileNamePrefix and ClearProfilePrefixByType change what the profile
	// shows in one season. The other seasons keep their selection.
	SelectProfileNameColor(ctx context.Context, queries sql.Queries, profileID, seasonID, nameColorID uuid.UUID) error
	SelectProfileNamePrefix(ctx context.Context, queries sql.Queries, profileID, seasonID, namePrefixID uuid.UUID, prefixType domain.ProfilePrefixType) error
	ClearProfilePrefixByType(ctx context.Context, queries sql.Queries, profileID, seasonID uuid.UUID, prefixType domain.ProfilePrefixType) error
	// PruneProfileSelections resets every selection of the profile, in every season, that no unrevoked option covers.
	// Call it after options are revoked.
	PruneProfileSelections(ctx context.Context, queries sql.Queries, profileID uuid.UUID) error
	AddProfileNameColorOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, nameColorID uuid.UUID, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error
	AddProfileNameGlythOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, namePrefixID uuid.UUID, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error
	AddProfileNamePrefixOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, namePrefixID uuid.UUID, prefixType domain.ProfilePrefixType, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error
	GetProfileNameColorOptionsByProfileOwnerUserID(ctx context.Context, ownerUserID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error)
	GetProfileNamePrefixOptionsByProfileOwnerUserIDAndType(ctx context.Context, ownerUserID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error)
	// GetProfilesCosmetics returns what every profile shows in the season, keyed by profile ID.
	// A profile without a selection gets the default name color and no prefixes.
	GetProfilesCosmetics(ctx context.Context, profileIDs uuid.UUIDs, seasonID uuid.UUID) (map[uuid.UUID]*domain.ProfileCosmetics, error)
	GetProfileFullPrefix(ctx context.Context, nameColor *domain.NameColor, glythPrefix *domain.NamePrefix, specialPrefix *domain.NamePrefix) string
	// GetProfileChatPrefix builds the chat prefix the season server gets from the selection of the profile in the season.
	GetProfileChatPrefix(ctx context.Context, profileID, seasonID uuid.UUID) (string, error)
	// GetProfileChatPrefixWithQueries is GetProfileChatPrefix reading through queries,
	// so it sees a selection written in the same transaction.
	GetProfileChatPrefixWithQueries(ctx context.Context, queries sql.Queries, profileID, seasonID uuid.UUID) (string, error)
}

type profileCosmeticsService struct {
	storage storage.MainStorage
}

func NewProfileCosmeticsService(
	storage storage.MainStorage,
) ProfileCosmeticsService {
	return &profileCosmeticsService{
		storage: storage,
	}
}

func (s *profileCosmeticsService) GetProfileNameColorOptions(ctx context.Context, profileID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error) {
	nameColorOptions, err := s.storage.Queries().FindProfileNameColorOptionsByProfileID(ctx, profileID, seasonID)
	if err == stdsql.ErrNoRows {
		return []*domain.ProfileNameColorOption{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile name color options", err)
	}
	return nameColorOptions, nil
}

func (s *profileCosmeticsService) GetProfileNamePrefixOptionsByProfileIDAndType(ctx context.Context, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error) {
	namePrefixOptions, err := s.storage.Queries().FindProfileNamePrefixOptionsByProfileIDAndType(ctx, profileID, prefixType, seasonID)
	if err == stdsql.ErrNoRows {
		return []*domain.ProfileNamePrefixOption{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile name prefix options by profile id and type", err)
	}
	return namePrefixOptions, nil
}

func (s *profileCosmeticsService) GetProfileNameColorOptionByIDAndProfileID(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, seasonID *uuid.UUID) (*domain.ProfileNameColorOption, error) {
	nameColorOption, err := s.storage.Queries().FindProfileNameColorOptionByIDAndProfileID(ctx, optionID, profileID, seasonID)
	if err == stdsql.ErrNoRows {
		return nil, utils.NewNotFoundError("profile name color option not found", nil)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile name color option by id", err)
	}
	return nameColorOption, nil
}

func (s *profileCosmeticsService) GetProfileNamePrefixOptionByIDAndProfileIDAndType(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) (*domain.ProfileNamePrefixOption, error) {
	namePrefixOption, err := s.storage.Queries().FindProfileNamePrefixOptionByIDAndProfileIDAndType(ctx, optionID, profileID, prefixType, seasonID)
	if err == stdsql.ErrNoRows {
		return nil, utils.NewNotFoundError("profile name prefix option not found", nil)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile name prefix option by id and profile id and type", err)
	}
	return namePrefixOption, nil
}

func (s *profileCosmeticsService) SelectProfileNameColor(ctx context.Context, queries sql.Queries, profileID, seasonID, nameColorID uuid.UUID) error {
	if err := queries.SetProfileSeasonNameColor(ctx, profileID, seasonID, nameColorID); err != nil {
		return utils.NewInternalServerError("failed to select profile name color", err)
	}
	return nil
}

func (s *profileCosmeticsService) SelectProfileNamePrefix(ctx context.Context, queries sql.Queries, profileID, seasonID, namePrefixID uuid.UUID, prefixType domain.ProfilePrefixType) error {
	if err := queries.SetProfileSeasonPrefix(ctx, profileID, seasonID, prefixType, &namePrefixID); err != nil {
		return utils.NewInternalServerError("failed to select profile name prefix by type "+string(prefixType), err)
	}
	return nil
}

func (s *profileCosmeticsService) ClearProfilePrefixByType(ctx context.Context, queries sql.Queries, profileID, seasonID uuid.UUID, prefixType domain.ProfilePrefixType) error {
	if err := queries.SetProfileSeasonPrefix(ctx, profileID, seasonID, prefixType, nil); err != nil {
		return utils.NewInternalServerError("failed to clear profile prefix by type "+string(prefixType), err)
	}
	return nil
}

func (s *profileCosmeticsService) PruneProfileSelections(ctx context.Context, queries sql.Queries, profileID uuid.UUID) error {
	if err := queries.PruneProfileSeasonCosmetics(ctx, profileID); err != nil {
		return utils.NewInternalServerError("failed to reset profile cosmetics selection", err)
	}
	return nil
}

// errPermanentPurchase means an option that comes from an order has no season. Only an admin grant is permanent.
var errPermanentPurchase = errors.New("a purchased cosmetic option must be for a season")

func (s *profileCosmeticsService) AddProfileNameColorOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, nameColorID uuid.UUID, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error {
	if forSeasonID == nil && orderItemID != nil {
		return utils.NewInternalServerError("failed to add profile name color option", errPermanentPurchase)
	}
	err := queries.InsertProfileNameColorOption(ctx, sql.InsertProfileNameColorOptionParams{
		ProfileID:   profileID,
		NameColorID: nameColorID,
		ForSeasonID: forSeasonID,
		OrderItemID: orderItemID,
		CreatedBy:   utils.GetUserIDFromContextOrNil(ctx),
	})
	if err != nil {
		return utils.NewInternalServerError("failed to add profile name color option", err)
	}
	return nil
}

func (s *profileCosmeticsService) AddProfileNameGlythOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, namePrefixID uuid.UUID, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error {
	return s.AddProfileNamePrefixOption(ctx, queries, profileID, namePrefixID, domain.ProfilePrefixTypeGlyth, forSeasonID, orderItemID)
}

func (s *profileCosmeticsService) AddProfileNamePrefixOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, namePrefixID uuid.UUID, prefixType domain.ProfilePrefixType, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error {
	if forSeasonID == nil && orderItemID != nil {
		return utils.NewInternalServerError("failed to add profile name "+string(prefixType)+" option", errPermanentPurchase)
	}
	err := queries.InsertProfileNamePrefixOption(ctx, sql.InsertProfileNamePrefixOptionParams{
		ProfileID:    profileID,
		NamePrefixID: namePrefixID,
		Type:         prefixType,
		ForSeasonID:  forSeasonID,
		OrderItemID:  orderItemID,
		CreatedBy:    utils.GetUserIDFromContextOrNil(ctx),
	})
	if err != nil {
		return utils.NewInternalServerError("failed to add profile name "+string(prefixType)+" option", err)
	}
	return nil
}

func (s *profileCosmeticsService) GetProfileNameColorOptionsByProfileOwnerUserID(ctx context.Context, ownerUserID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error) {
	nameColorOptions, err := s.storage.Queries().FindProfileNameColorOptionsByProfileOwnerUserID(ctx, ownerUserID, seasonID)
	if err == stdsql.ErrNoRows {
		return []*domain.ProfileNameColorOption{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile name color options by profile owner user id", err)
	}
	return nameColorOptions, nil
}

func (s *profileCosmeticsService) GetProfileNamePrefixOptionsByProfileOwnerUserIDAndType(ctx context.Context, ownerUserID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error) {
	namePrefixOptions, err := s.storage.Queries().FindProfileNamePrefixOptionsByProfileOwnerUserIDAndType(ctx, ownerUserID, prefixType, seasonID)
	if err == stdsql.ErrNoRows {
		return []*domain.ProfileNamePrefixOption{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile name prefix options by profile owner user id and type", err)
	}
	return namePrefixOptions, nil
}

func (s *profileCosmeticsService) GetProfilesCosmetics(ctx context.Context, profileIDs uuid.UUIDs, seasonID uuid.UUID) (map[uuid.UUID]*domain.ProfileCosmetics, error) {
	return s.findCosmetics(ctx, s.storage.Queries(), profileIDs, seasonID)
}

func (s *profileCosmeticsService) findCosmetics(ctx context.Context, queries sql.Queries, profileIDs uuid.UUIDs, seasonID uuid.UUID) (map[uuid.UUID]*domain.ProfileCosmetics, error) {
	cosmetics, err := queries.FindProfilesSeasonCosmetics(ctx, profileIDs, seasonID, config.GetDefaultNameColorID())
	if err != nil {
		// Callers fall back to empty cosmetics and log the message only, so the cause is logged here.
		wrapped := utils.NewInternalServerError("failed to get profiles cosmetics", err)
		utils.LogCustomError(ctx, wrapped)
		return nil, wrapped
	}
	return cosmetics, nil
}

func (s *profileCosmeticsService) GetProfileChatPrefix(ctx context.Context, profileID, seasonID uuid.UUID) (string, error) {
	return s.GetProfileChatPrefixWithQueries(ctx, s.storage.Queries(), profileID, seasonID)
}

func (s *profileCosmeticsService) GetProfileChatPrefixWithQueries(ctx context.Context, queries sql.Queries, profileID, seasonID uuid.UUID) (string, error) {
	cosmetics, err := s.findCosmetics(ctx, queries, uuid.UUIDs{profileID}, seasonID)
	if err != nil {
		return "", err
	}
	selected, ok := cosmetics[profileID]
	if !ok {
		return "", utils.NewNotFoundError("profile not found", nil)
	}
	return s.GetProfileFullPrefix(ctx, selected.NameColor, selected.Glyth, selected.Special), nil
}

func (s *profileCosmeticsService) GetProfileFullPrefix(ctx context.Context, nameColor *domain.NameColor, glythPrefix *domain.NamePrefix, specialPrefix *domain.NamePrefix) string {
	var colors string
	if nameColor == nil || len(nameColor.Metadata.Colors) == 0 {
		colors = ""
	} else {
		colors = utils.FormatNameColorsToMiniMessages(ctx, nameColor.Metadata.Colors)
	}

	var glythPrefixString string
	if glythPrefix == nil || glythPrefix.Metadata.Prefix == "" {
		glythPrefixString = ""
	} else {
		glythPrefixString = glythPrefix.Metadata.Prefix
		if !glythPrefix.Metadata.NoSpace {
			glythPrefixString = glythPrefixString + " "
		}
	}

	var specialPrefixString string
	if specialPrefix == nil || specialPrefix.Metadata.Prefix == "" {
		specialPrefixString = ""
	} else {
		specialPrefixString = specialPrefix.Metadata.Prefix
		if !specialPrefix.Metadata.NoSpace {
			specialPrefixString = specialPrefixString + " "
		}
	}
	prefix := fmt.Sprintf("<reset>%s%s%s", specialPrefixString, glythPrefixString, colors)
	return prefix
}
