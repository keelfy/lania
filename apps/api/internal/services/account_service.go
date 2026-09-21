package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

// AccountDeletionMaxAuthAge is how long after a sign in the account can be deleted.
// It matches selfservice.flows.settings.privileged_session_max_age in kratos.yml.
const AccountDeletionMaxAuthAge = 15 * time.Minute

// ErrSessionRefreshRequired is the message of the forbidden error for a sign in older than AccountDeletionMaxAuthAge.
// The client signs the user in again when it gets it.
const ErrSessionRefreshRequired = "session_refresh_required"

type AccountService interface {
	// DeleteAccount deletes the Ory identity of the user. Before that it releases every profile of the user:
	// cosmetics are revoked, the role becomes player and the profile gets no owner, so anybody can claim it.
	// Access to seasons, stats and orders stay. The basket and the notifications of the user are deleted.
	// The owner of the site cannot delete the account. A failed request can be repeated.
	DeleteAccount(ctx context.Context, userID uuid.UUID, siteRole domain.Role, authenticatedAt time.Time) error
}

type accountService struct {
	storage                 storage.MainStorage
	oryClient               clients.OryAPI
	profileService          ProfileService
	profileCosmeticsService ProfileCosmeticsService
	profileResyncService    ProfileResyncService

	now func() time.Time
}

func NewAccountService(
	storage storage.MainStorage,
	oryClient clients.OryAPI,
	profileService ProfileService,
	profileCosmeticsService ProfileCosmeticsService,
	profileResyncService ProfileResyncService,
) AccountService {
	return &accountService{
		storage:                 storage,
		oryClient:               oryClient,
		profileService:          profileService,
		profileCosmeticsService: profileCosmeticsService,
		profileResyncService:    profileResyncService,
		now:                     time.Now,
	}
}

func (s *accountService) DeleteAccount(ctx context.Context, userID uuid.UUID, siteRole domain.Role, authenticatedAt time.Time) error {
	if siteRole == domain.RoleOwner {
		return utils.NewConflictError("the owner of the site cannot delete the account", nil)
	}
	if s.now().Sub(authenticatedAt) > AccountDeletionMaxAuthAge {
		return utils.NewForbiddenError(ErrSessionRefreshRequired, nil)
	}

	profiles, err := s.profileService.GetProfilesByOwnerUserID(ctx, userID)
	if err != nil {
		return err
	}

	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		for _, profile := range profiles {
			if err := s.releaseProfile(ctx, queries, profile, userID); err != nil {
				return err
			}
		}
		if err := queries.ClearBasketItemsByUserID(ctx, userID); err != nil {
			return utils.NewInternalServerError("failed to clear the basket", err)
		}
		if err := queries.DeleteNotificationsByUserID(ctx, userID); err != nil {
			return utils.NewInternalServerError("failed to delete notifications", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Every season server is called, which can be slow, so the request does not wait for it.
	// Role sync pushes the role again anyway, and an admin can resync a profile by hand.
	resyncCtx := context.WithoutCancel(ctx)
	go func() {
		for _, profile := range profiles {
			if err := s.profileResyncService.ResyncProfile(resyncCtx, profile.ID); err != nil {
				logger.Warnf(resyncCtx, "failed to resync profile %s released by a deleted account: %v", profile.ID, err)
			}
		}
	}()

	// The identity goes last: until it is deleted, the user can sign in and repeat the request.
	err = s.oryClient.DeleteIdentity(ctx, userID.String())
	if err != nil && !errors.Is(err, clients.ErrIdentityNotFound) {
		return utils.NewInternalServerError("profiles are released, but the account was not deleted, repeat the request to retry", err)
	}
	return nil
}

// releaseProfile takes back what the user got for the profile and leaves the profile without an owner.
func (s *accountService) releaseProfile(ctx context.Context, queries sql.Queries, profile *domain.Profile, userID uuid.UUID) error {
	defaultNameColorID := config.GetDefaultNameColorID()

	// Every profile starts with the default color, so it stays.
	if err := queries.RevokeProfileCosmetics(ctx, profile.ID, defaultNameColorID, &userID); err != nil {
		return utils.NewInternalServerError("failed to revoke profile cosmetics", err)
	}
	if err := s.profileCosmeticsService.SelectProfileNameColor(ctx, queries, profile.ID, defaultNameColorID); err != nil {
		return err
	}
	for _, prefixType := range []domain.ProfilePrefixType{domain.ProfilePrefixTypeGlyth, domain.ProfilePrefixTypeSpecial} {
		if err := s.profileCosmeticsService.ClearProfilePrefixByType(ctx, queries, profile.ID, prefixType); err != nil {
			return err
		}
	}

	if profile.Role != domain.RolePlayer {
		if err := queries.SetProfileRole(ctx, profile.ID, domain.RolePlayer, &userID); err != nil {
			return utils.NewInternalServerError("failed to reset profile role", err)
		}
	}
	if err := queries.SetProfileOwner(ctx, profile.ID, nil, &userID); err != nil {
		return utils.NewInternalServerError("failed to release profile", err)
	}
	return nil
}
