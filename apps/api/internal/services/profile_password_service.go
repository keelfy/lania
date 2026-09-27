package services

import (
	"context"
	stdsql "database/sql"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// ErrPlayerNotRegisteredMessage is the API message for a profile that never registered in game, so the site can
// tell the player to join and register instead.
const ErrPlayerNotRegisteredMessage = "player_not_registered"

const (
	inGamePasswordMinLength = 5
	inGamePasswordMaxLength = 32
	// bcryptMaxBytes is the longest input bcrypt hashes.
	bcryptMaxBytes = 72
)

// ProfilePasswordService gives the owner of an unlicensed profile a new in-game password, for a forgotten password
// or for someone who registered in game under the nickname before the owner did.
type ProfilePasswordService interface {
	// SetOwnedPassword replaces the in-game password of the profile in the season, turns off its two-factor login
	// and kicks the player from the server. The profile must be unlicensed and have access to the season, and the season must be active.
	SetOwnedPassword(ctx context.Context, userID, profileID, seasonID uuid.UUID, password string) error
}

type profilePasswordService struct {
	storage          storage.MainStorage
	accessService    AccessService
	seasonService    SeasonService
	minecraftService MinecraftService
}

func NewProfilePasswordService(
	storage storage.MainStorage,
	accessService AccessService,
	seasonService SeasonService,
	minecraftService MinecraftService,
) ProfilePasswordService {
	return &profilePasswordService{
		storage:          storage,
		accessService:    accessService,
		seasonService:    seasonService,
		minecraftService: minecraftService,
	}
}

func (s *profilePasswordService) SetOwnedPassword(ctx context.Context, userID, profileID, seasonID uuid.UUID, password string) error {
	if err := validateInGamePassword(password); err != nil {
		return err
	}

	profile, err := s.storage.Queries().FindProfileByID(ctx, profileID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return utils.NewNotFoundError("profile not found", err)
	} else if err != nil {
		return utils.NewInternalServerError("failed to get profile", err)
	}
	if profile.OwnerUserID == nil || *profile.OwnerUserID != userID {
		return utils.NewForbiddenError("only owner can change the in-game password of the profile", nil)
	}

	mojangUUIDs, err := s.storage.Queries().FindProfileMojangUUIDsByMinecraftUUIDs(ctx, uuid.UUIDs{profile.MinecraftUUID})
	if err != nil {
		return utils.NewInternalServerError("failed to get the mojang uuid of the profile", err)
	}
	if mojangUUIDs[profile.MinecraftUUID] == profile.MinecraftUUID {
		return utils.NewConflictError("a licensed profile logs in without a password", nil)
	}

	season, err := s.seasonService.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return err
	}
	if !season.IsActive {
		return utils.NewBadRequestError("the in-game password can be changed in an active season only", nil)
	}
	hasAccess, err := s.accessService.CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, profile.MinecraftUUID, seasonID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return utils.NewForbiddenError("the profile has no access to the season", nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return utils.NewInternalServerError("failed to hash the password", err)
	}
	err = s.minecraftService.SetPasswordInSeason(ctx, seasonID, profile, string(hash))
	if errors.Is(err, ErrPlayerNotRegistered) {
		return utils.NewConflictError(ErrPlayerNotRegisteredMessage, err)
	}
	return err
}

// validateInGamePassword follows the default rules of NavAuth, so the password also passes them in game. The
// length cap keeps most passwords within the 72 bytes bcrypt hashes.
func validateInGamePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if !utf8.ValidString(password) || length < inGamePasswordMinLength || length > inGamePasswordMaxLength || len(password) > bcryptMaxBytes {
		return utils.NewBadRequestError("a password is 5 to 32 characters", nil)
	}
	if strings.IndexFunc(password, func(r rune) bool { return unicode.IsSpace(r) || !unicode.IsPrint(r) }) >= 0 {
		return utils.NewBadRequestError("a password has no spaces", nil)
	}
	if !strings.ContainsFunc(password, unicode.IsLower) || !strings.ContainsFunc(password, unicode.IsDigit) {
		return utils.NewBadRequestError("a password has a lowercase letter and a digit", nil)
	}
	return nil
}
