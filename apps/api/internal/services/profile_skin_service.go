package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"regexp"
	"sync"
	"time"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

// MaxSkinFileBytes bounds a skin file. A 64x64 PNG is a few kilobytes.
const MaxSkinFileBytes = 64 * 1024

// ownerSkinCooldown is the time an owner waits between two skin changes of the same profile. Every skin given by a
// file costs a MineSkin request, and MineSkin has a rate limit.
const ownerSkinCooldown = time.Minute

// Messages the site tells apart.
const (
	ErrSkinFileInvalidMessage         = "skin_file_invalid"
	ErrSkinNicknameNotLicensedMessage = "skin_nickname_not_licensed"
	ErrSkinCooldownMessage            = "skin_cooldown"
)

var skinFileIDPattern = regexp.MustCompile(`^[0-9a-f]{16}$`)

// ProfileSkinService lets the owner of a profile change the skin the player wears on the server of a season. The
// server keeps the skin with SkinsRestorer, the same way as one set in game.
type ProfileSkinService interface {
	// SetOwnedSkinFile gives the profile the skin of a 64x64 or 64x32 PNG file. The file is stored, so the server
	// can download it. It fails with ErrSkinPending when the server does not apply the skin in time.
	SetOwnedSkinFile(ctx context.Context, userID, profileID, seasonID uuid.UUID, content []byte, variant domain.SkinVariant) (*domain.PlayerSkin, error)
	// SetOwnedSkinNickname gives the profile the skin of the licensed account with the nickname. It fails with
	// ErrSkinPending when the server does not apply the skin in time.
	SetOwnedSkinNickname(ctx context.Context, userID, profileID, seasonID uuid.UUID, nickname string) (*domain.PlayerSkin, error)
	// ClearOwnedSkin brings back the skin of the licensed account of the profile, or the default one. It fails
	// with ErrSkinPending when the server does not do it in time.
	ClearOwnedSkin(ctx context.Context, userID, profileID, seasonID uuid.UUID) error
	// GetSkinFile returns a stored skin file. It fails with a not found error for an unknown file.
	GetSkinFile(ctx context.Context, fileID string) ([]byte, error)
}

type profileSkinService struct {
	profileService   ProfileService
	accessService    AccessService
	seasonService    SeasonService
	minecraftService MinecraftService
	mojangService    MojangService
	objects          clients.ObjectStorage

	now func() time.Time
	// lastChange holds the time of the last skin change per profile. It lives in memory, so every API instance
	// keeps its own cooldown.
	mu         sync.Mutex
	lastChange map[uuid.UUID]time.Time
}

func NewProfileSkinService(
	profileService ProfileService,
	accessService AccessService,
	seasonService SeasonService,
	minecraftService MinecraftService,
	mojangService MojangService,
	objects clients.ObjectStorage,
) ProfileSkinService {
	return &profileSkinService{
		profileService:   profileService,
		accessService:    accessService,
		seasonService:    seasonService,
		minecraftService: minecraftService,
		mojangService:    mojangService,
		objects:          objects,
		now:              time.Now,
		lastChange:       make(map[uuid.UUID]time.Time),
	}
}

func (s *profileSkinService) SetOwnedSkinFile(ctx context.Context, userID, profileID, seasonID uuid.UUID, content []byte, variant domain.SkinVariant) (*domain.PlayerSkin, error) {
	if variant != domain.SkinVariantClassic && variant != domain.SkinVariantSlim {
		return nil, utils.NewBadRequestError("variant must be classic or slim", nil)
	}
	if err := validateSkinFile(content); err != nil {
		return nil, err
	}
	publicURL := config.GetApiPublicURL()
	if publicURL == "" {
		return nil, utils.NewInternalServerError("API_PUBLIC_URL is not set, the server cannot download skin files", nil)
	}
	profile, err := s.ownedProfileInSeason(ctx, userID, profileID, seasonID)
	if err != nil {
		return nil, err
	}

	// The same file gets the same link, so SkinsRestorer reuses the skin it generated for it before.
	hash := sha256.Sum256(content)
	fileID := hex.EncodeToString(hash[:])[:16]
	if err := s.objects.PutObject(ctx, skinFileKey(fileID), content, "image/png"); err != nil {
		return nil, utils.NewInternalServerError("failed to store the skin file", err)
	}

	if err := s.startChange(profileID); err != nil {
		return nil, err
	}
	link := fmt.Sprintf("%s/v1/skins/files/%s.png", publicURL, fileID)
	return s.setSkin(ctx, seasonID, profile, link, variant, nil)
}

func (s *profileSkinService) SetOwnedSkinNickname(ctx context.Context, userID, profileID, seasonID uuid.UUID, nickname string) (*domain.PlayerSkin, error) {
	if !mojangUsernameRegexp.MatchString(nickname) {
		return nil, utils.NewBadRequestError("nickname is 3 to 16 letters, digits or underscores", nil)
	}
	profile, err := s.ownedProfileInSeason(ctx, userID, profileID, seasonID)
	if err != nil {
		return nil, err
	}
	_, mojangUUID, err := s.mojangService.ResolveGameUUID(ctx, nickname)
	if err != nil {
		return nil, err
	}
	if mojangUUID == nil {
		return nil, utils.NewBadRequestError(ErrSkinNicknameNotLicensedMessage, nil)
	}

	if err := s.startChange(profileID); err != nil {
		return nil, err
	}
	return s.setSkin(ctx, seasonID, profile, nickname, "", mojangUUID)
}

func (s *profileSkinService) ClearOwnedSkin(ctx context.Context, userID, profileID, seasonID uuid.UUID) error {
	profile, err := s.ownedProfileInSeason(ctx, userID, profileID, seasonID)
	if err != nil {
		return err
	}
	if err := s.startChange(profileID); err != nil {
		return err
	}
	// Waits past the request timeout, like setSkin.
	return s.minecraftService.ClearPlayerSkinInSeason(context.WithoutCancel(ctx), seasonID, profile)
}

func (s *profileSkinService) GetSkinFile(ctx context.Context, fileID string) ([]byte, error) {
	if !skinFileIDPattern.MatchString(fileID) {
		return nil, utils.NewNotFoundError("skin file not found", nil)
	}
	content, err := s.objects.GetObject(ctx, skinFileKey(fileID))
	// S3-compatible storages do not all name the error NoSuchKey, the status is the same.
	var responseErr *awshttp.ResponseError
	if errors.As(err, &responseErr) && responseErr.HTTPStatusCode() == http.StatusNotFound {
		return nil, utils.NewNotFoundError("skin file not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to read the skin file", err)
	}
	return content, nil
}

// setSkin waits for the server past the request timeout: the server applies the skin anyway, so the caller is
// better off with the result. The shell client bounds the wait.
func (s *profileSkinService) setSkin(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile, skin string, variant domain.SkinVariant, mojangUUID *uuid.UUID) (*domain.PlayerSkin, error) {
	return s.minecraftService.SetPlayerSkinInSeason(context.WithoutCancel(ctx), seasonID, profile, skin, variant, mojangUUID)
}

// ownedProfileInSeason returns the profile when the user owns it and it can play in the season, which is active.
func (s *profileSkinService) ownedProfileInSeason(ctx context.Context, userID, profileID, seasonID uuid.UUID) (*domain.Profile, error) {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if profile.OwnerUserID == nil || *profile.OwnerUserID != userID {
		return nil, utils.NewForbiddenError("only owner can change the skin of the profile", nil)
	}

	season, err := s.seasonService.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	if !season.IsActive {
		return nil, utils.NewBadRequestError("the skin can be changed in an active season only", nil)
	}
	hasAccess, err := s.accessService.CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, profile.MinecraftUUID, seasonID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, utils.NewForbiddenError("the profile has no access to the season", nil)
	}
	return profile, nil
}

// startChange starts the cooldown of the profile, or fails with a too many requests error while it runs.
func (s *profileSkinService) startChange(profileID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for id, last := range s.lastChange {
		if now.Sub(last) >= ownerSkinCooldown {
			delete(s.lastChange, id)
		}
	}
	if _, waiting := s.lastChange[profileID]; waiting {
		return utils.NewTooManyRequestsError(ErrSkinCooldownMessage, nil)
	}
	s.lastChange[profileID] = now
	return nil
}

// validateSkinFile accepts what SkinsRestorer and the game accept: a whole PNG of 64x64, or 64x32 in the legacy
// format. HD skins are refused.
func validateSkinFile(content []byte) error {
	if len(content) == 0 || len(content) > MaxSkinFileBytes {
		return utils.NewBadRequestError(ErrSkinFileInvalidMessage, nil)
	}
	decoded, err := png.Decode(bytes.NewReader(content))
	if err != nil {
		return utils.NewBadRequestError(ErrSkinFileInvalidMessage, err)
	}
	if size := decoded.Bounds().Size(); size.X != 64 || (size.Y != 64 && size.Y != 32) {
		return utils.NewBadRequestError(ErrSkinFileInvalidMessage, nil)
	}
	return nil
}

func skinFileKey(fileID string) string {
	return "skins/" + fileID + ".png"
}
