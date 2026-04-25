package services

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"github.com/google/uuid"
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
	SelectProfileNameColor(ctx context.Context, queries sql.Queries, profileID uuid.UUID, nameColorID uuid.UUID) error
	SelectProfileNamePrefix(ctx context.Context, queries sql.Queries, profileID uuid.UUID, namePrefixID uuid.UUID, prefixType domain.ProfilePrefixType) error
	ClearProfilePrefixByType(ctx context.Context, queries sql.Queries, profileID uuid.UUID, prefixType domain.ProfilePrefixType) error
	AddProfileNameColorOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, nameColorID uuid.UUID, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error
	AddProfileNameGlythOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, namePrefixID uuid.UUID, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error
	GetProfileNameColorOptionsByProfileOwnerUserID(ctx context.Context, ownerUserID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error)
	GetProfileNamePrefixOptionsByProfileOwnerUserIDAndType(ctx context.Context, ownerUserID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error)
	GetProfilePrefixes(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfilePrefix, error)
	GetProfileFullPrefix(ctx context.Context, nameColor *domain.NameColor, glythPrefix *domain.NamePrefix, specialPrefix *domain.NamePrefix) string
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

func (s *profileCosmeticsService) SelectProfileNameColor(ctx context.Context, queries sql.Queries, profileID uuid.UUID, nameColorID uuid.UUID) error {
	err := queries.UpdateProfileNameColorByID(ctx, profileID, nameColorID)
	if err != nil {
		return utils.NewInternalServerError("failed to select profile name color", err)
	}
	return nil
}

func (s *profileCosmeticsService) SelectProfileNamePrefix(ctx context.Context, queries sql.Queries, profileID uuid.UUID, namePrefixID uuid.UUID, prefixType domain.ProfilePrefixType) error {
	prefixes, err := queries.FindProfilePrefixesByProfileID(ctx, profileID)
	if err != nil {
		return utils.NewInternalServerError("failed to find profile prefixes", err)
	}

	var existingPrefix *domain.ProfilePrefix
	for _, p := range prefixes {
		if p.Type == prefixType {
			existingPrefix = p
			break
		}
	}

	if existingPrefix != nil {
		err = queries.UpdateProfileNamePrefixByProfileIDAndType(ctx, profileID, namePrefixID, prefixType)
	} else {
		err = queries.InsertProfilePrefix(ctx, sql.InsertProfilePrefixParams{
			ProfileID:    profileID,
			NamePrefixID: namePrefixID,
			Type:         prefixType,
		})
	}
	if err != nil {
		return utils.NewInternalServerError("failed to select profile name prefix by type "+string(prefixType), err)
	}

	return nil
}

func (s *profileCosmeticsService) ClearProfilePrefixByType(ctx context.Context, queries sql.Queries, profileID uuid.UUID, prefixType domain.ProfilePrefixType) error {
	prefixes, err := queries.FindProfilePrefixesByProfileID(ctx, profileID)
	if err != nil {
		return utils.NewInternalServerError("failed to find profile prefixes", err)
	}

	var existingPrefix *domain.ProfilePrefix
	for _, p := range prefixes {
		if p.Type == prefixType {
			existingPrefix = p
			break
		}
	}

	if existingPrefix != nil {
		err = queries.DeleteProfilePrefixByProfileIDAndType(ctx, profileID, prefixType)
		if err != nil {
			return utils.NewInternalServerError("failed to clear profile prefix by type "+string(prefixType), err)
		}
	}

	return nil
}

func (s *profileCosmeticsService) AddProfileNameColorOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, nameColorID uuid.UUID, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error {
	err := queries.InsertProfileNameColorOption(ctx, sql.InsertProfileNameColorOptionParams{
		ProfileID:   profileID,
		NameColorID: nameColorID,
		ForSeasonID: forSeasonID,
		OrderItemID: orderItemID,
	})
	if err != nil {
		return utils.NewInternalServerError("failed to add profile name color option", err)
	}
	return nil
}

func (s *profileCosmeticsService) AddProfileNameGlythOption(ctx context.Context, queries sql.Queries, profileID uuid.UUID, namePrefixID uuid.UUID, forSeasonID *uuid.UUID, orderItemID *uuid.UUID) error {
	err := queries.InsertProfileNamePrefixOption(ctx, sql.InsertProfileNamePrefixOptionParams{
		ProfileID:    profileID,
		NamePrefixID: namePrefixID,
		Type:         domain.ProfilePrefixTypeGlyth,
		ForSeasonID:  forSeasonID,
		OrderItemID:  orderItemID,
	})
	if err != nil {
		return utils.NewInternalServerError("failed to add profile name glyth option", err)
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

func (s *profileCosmeticsService) GetProfilePrefixes(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfilePrefix, error) {
	prefixes, err := s.storage.Queries().FindProfilePrefixesByProfileID(ctx, profileID)
	if err == stdsql.ErrNoRows {
		return []*domain.ProfilePrefix{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile prefixes", err)
	}
	return prefixes, nil
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
