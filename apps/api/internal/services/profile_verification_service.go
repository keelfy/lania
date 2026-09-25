package services

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	stdsql "database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

const (
	// verificationTTL is how long a request waits for the player to join, and then for the owner to type the code.
	verificationTTL = 10 * time.Minute
	// verificationMaxAttempts wrong codes close the request.
	verificationMaxAttempts = 5
	verificationCodeLength  = 6
	// verificationCodeAlphabet has no characters that look alike on the kick screen (0/O, 1/I/L).
	verificationCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
)

// ProfileVerificationService lets the owner prove that the licensed account of the profile is theirs.
// The proxy plugin shows a code to the licensed player on join, the owner types it on the site.
type ProfileVerificationService interface {
	// StartVerification opens a request for the owner. The profile must be keyed to its Mojang UUID.
	// It returns when the request expires unless the player joins.
	StartVerification(ctx context.Context, userID, profileID uuid.UUID) (time.Time, error)
	// IssueLoginCode is asked by the proxy on every login. It returns the code to show when the player has an open
	// request and joined in online mode with the UUID of the profile, and an empty string otherwise.
	IssueLoginCode(ctx context.Context, mcUUID uuid.UUID, username string, onlineMode bool) (string, error)
	// ConfirmVerification checks the code the owner typed. A wrong code is a bad request error; after
	// verificationMaxAttempts wrong codes, or once the request expired, it is a conflict error.
	ConfirmVerification(ctx context.Context, userID, profileID uuid.UUID, code string) error
}

type profileVerificationService struct {
	storage storage.MainStorage
}

func NewProfileVerificationService(storage storage.MainStorage) ProfileVerificationService {
	return &profileVerificationService{storage: storage}
}

func (s *profileVerificationService) ownedProfile(ctx context.Context, userID, profileID uuid.UUID) (*domain.Profile, error) {
	profile, err := s.storage.Queries().FindProfileByID(ctx, profileID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewNotFoundError("profile not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile", err)
	}
	if profile.OwnerUserID == nil || *profile.OwnerUserID != userID {
		return nil, utils.NewForbiddenError("only owner can verify the profile", nil)
	}
	return profile, nil
}

func (s *profileVerificationService) StartVerification(ctx context.Context, userID, profileID uuid.UUID) (time.Time, error) {
	profile, err := s.ownedProfile(ctx, userID, profileID)
	if err != nil {
		return time.Time{}, err
	}
	if profile.VerifiedAt != nil {
		return time.Time{}, utils.NewConflictError("the profile is already verified", nil)
	}

	queries := s.storage.Queries()
	mojangUUIDs, err := queries.FindProfileMojangUUIDsByMinecraftUUIDs(ctx, uuid.UUIDs{profile.MinecraftUUID})
	if err != nil {
		return time.Time{}, utils.NewInternalServerError("failed to get the mojang uuid of the profile", err)
	}
	if mojangUUIDs[profile.MinecraftUUID] != profile.MinecraftUUID {
		return time.Time{}, utils.NewBadRequestError("only a licensed profile can be verified", nil)
	}

	if err := queries.UpsertProfileVerification(ctx, profileID, verificationTTL); err != nil {
		return time.Time{}, utils.NewInternalServerError("failed to open the verification", err)
	}
	verification, err := queries.FindOpenProfileVerification(ctx, profileID)
	if err != nil {
		return time.Time{}, utils.NewInternalServerError("failed to get the verification", err)
	}
	return verification.ExpiresAt, nil
}

func (s *profileVerificationService) IssueLoginCode(ctx context.Context, mcUUID uuid.UUID, username string, onlineMode bool) (string, error) {
	queries := s.storage.Queries()
	profile, err := queries.FindProfileByUsername(ctx, username)
	if errors.Is(err, stdsql.ErrNoRows) {
		return "", nil
	} else if err != nil {
		return "", utils.NewInternalServerError("failed to get profile", err)
	}

	verification, err := queries.FindOpenProfileVerification(ctx, profile.ID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return "", nil
	} else if err != nil {
		return "", utils.NewInternalServerError("failed to get the verification", err)
	}

	// Only a login checked by Mojang proves the license; the UUID must be the one the profile is keyed to.
	if !onlineMode {
		return "", nil
	}
	if mcUUID != profile.MinecraftUUID {
		logger.Warnf(ctx, "verification of %s: joined as %s, the profile has %s; rekey the profile", username, mcUUID, profile.MinecraftUUID)
		return "", nil
	}

	code := ""
	if verification.Code != nil {
		code = *verification.Code
	} else if code, err = newVerificationCode(); err != nil {
		return "", utils.NewInternalServerError("failed to generate the verification code", err)
	}
	if err := queries.SetProfileVerificationCode(ctx, profile.ID, code, verificationTTL); err != nil {
		return "", utils.NewInternalServerError("failed to store the verification code", err)
	}
	return code, nil
}

func (s *profileVerificationService) ConfirmVerification(ctx context.Context, userID, profileID uuid.UUID, code string) error {
	profile, err := s.ownedProfile(ctx, userID, profileID)
	if err != nil {
		return err
	}

	queries := s.storage.Queries()
	verification, err := queries.FindOpenProfileVerification(ctx, profileID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return utils.NewConflictError("the verification expired, start it again", nil)
	} else if err != nil {
		return utils.NewInternalServerError("failed to get the verification", err)
	}
	// Before the player joined there is no code to guess, so the attempt is not counted.
	if verification.Code == nil {
		return utils.NewBadRequestError("wrong verification code", nil)
	}

	typed := strings.ToUpper(strings.TrimSpace(code))
	if subtle.ConstantTimeCompare([]byte(typed), []byte(*verification.Code)) != 1 {
		if verification.Attempts+1 >= verificationMaxAttempts {
			if err := queries.DeleteProfileVerification(ctx, profileID); err != nil {
				return utils.NewInternalServerError("failed to close the verification", err)
			}
			return utils.NewConflictError("too many wrong codes, start the verification again", nil)
		}
		if err := queries.IncrementProfileVerificationAttempts(ctx, profileID); err != nil {
			return utils.NewInternalServerError("failed to count the attempt", err)
		}
		return utils.NewBadRequestError("wrong verification code", nil)
	}

	if err := queries.SetProfileVerified(ctx, profileID, profile.MinecraftUUID); err != nil {
		return utils.NewInternalServerError("failed to verify the profile", err)
	}
	if err := queries.DeleteProfileVerification(ctx, profileID); err != nil {
		logger.Errorf(ctx, "failed to close the verification of %s: %v", profileID, err)
	}
	return nil
}

func newVerificationCode() (string, error) {
	random := make([]byte, verificationCodeLength)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	code := make([]byte, verificationCodeLength)
	for i, b := range random {
		code[i] = verificationCodeAlphabet[int(b)%len(verificationCodeAlphabet)]
	}
	return string(code), nil
}
