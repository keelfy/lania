package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/utils"
	ory "github.com/ory/client-go"
)

const (
	// identityScanPageSize and identityScanMaxPages bound a scan over all identities.
	identityScanPageSize = 250
	identityScanMaxPages = 40
)

type AdminUserService interface {
	// ListUsers returns a page of users and the token of the next page, empty on the last page.
	// A search query looks users up by email instead: it returns at most size users and no next page.
	ListUsers(ctx context.Context, search, pageToken string, size int) ([]*domain.User, string, error)
	// GetUserByID returns the user with the profiles the user owns.
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	// FindUserByEmail returns the user with exactly this email.
	// It fails with a not found error when nobody has the email and with a conflict error when several users have it.
	FindUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type adminUserService struct {
	oryClient      clients.OryAPI
	profileService ProfileService
}

func NewAdminUserService(
	oryClient clients.OryAPI,
	profileService ProfileService,
) AdminUserService {
	return &adminUserService{
		oryClient:      oryClient,
		profileService: profileService,
	}
}

func (s *adminUserService) ListUsers(ctx context.Context, search, pageToken string, size int) ([]*domain.User, string, error) {
	if search != "" {
		users, err := s.searchUsers(ctx, search, size)
		return users, "", err
	}

	identities, nextPageToken, err := s.oryClient.ListIdentities(ctx, int64(size), pageToken, "")
	if err != nil {
		return nil, "", utils.NewInternalServerError("failed to list identities", err)
	}
	return usersFromIdentities(identities), nextPageToken, nil
}

func (s *adminUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	identity, err := s.oryClient.GetIdentity(ctx, id.String())
	if errors.Is(err, clients.ErrIdentityNotFound) {
		return nil, utils.NewNotFoundError("user not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get identity", err)
	}

	user := userFromIdentity(identity)
	if user == nil {
		return nil, utils.NewInternalServerError("identity has an invalid id", nil)
	}

	user.Profiles, err = s.profileService.GetProfilesByOwnerUserID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *adminUserService) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return nil, utils.NewBadRequestError("email is required", nil)
	}

	users, err := s.lookupUsersByLoginIdentifier(ctx, email)
	if err != nil {
		return nil, err
	}
	// Users of a social login have no email as the login identifier, so only a scan finds them.
	if len(users) == 0 {
		users, err = s.scanUsers(ctx, 2, func(user *domain.User) bool {
			return normalizeEmail(user.Email) == email
		})
		if err != nil {
			return nil, err
		}
	}

	switch len(users) {
	case 0:
		return nil, utils.NewNotFoundError("user not found", nil)
	case 1:
		return users[0], nil
	}
	return nil, utils.NewConflictError("several users have this email", nil)
}

func (s *adminUserService) searchUsers(ctx context.Context, search string, limit int) ([]*domain.User, error) {
	query := normalizeEmail(search)

	users, err := s.lookupUsersByLoginIdentifier(ctx, query)
	if err != nil || len(users) > 0 {
		return users, err
	}
	return s.scanUsers(ctx, limit, func(user *domain.User) bool {
		return strings.Contains(normalizeEmail(user.Email), query)
	})
}

// lookupUsersByLoginIdentifier finds users that log in with exactly this identifier.
func (s *adminUserService) lookupUsersByLoginIdentifier(ctx context.Context, identifier string) ([]*domain.User, error) {
	identities, _, err := s.oryClient.ListIdentities(ctx, identityScanPageSize, "", identifier)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to look up identities by login identifier", err)
	}
	return usersFromIdentities(identities), nil
}

// scanUsers walks the identities page by page and collects the users that match,
// until it has limit users, runs out of identities or reaches identityScanMaxPages.
func (s *adminUserService) scanUsers(ctx context.Context, limit int, match func(*domain.User) bool) ([]*domain.User, error) {
	users := make([]*domain.User, 0, limit)
	pageToken := ""

	for range identityScanMaxPages {
		identities, nextPageToken, err := s.oryClient.ListIdentities(ctx, identityScanPageSize, pageToken, "")
		if err != nil {
			return nil, utils.NewInternalServerError("failed to scan identities", err)
		}

		for _, user := range usersFromIdentities(identities) {
			if !match(user) {
				continue
			}
			users = append(users, user)
			if len(users) == limit {
				return users, nil
			}
		}

		if nextPageToken == "" {
			return users, nil
		}
		pageToken = nextPageToken
	}

	logger.Warnf(ctx, "identity scan stopped after %d pages, results may be incomplete", identityScanMaxPages)
	return users, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func usersFromIdentities(identities []ory.Identity) []*domain.User {
	users := make([]*domain.User, 0, len(identities))
	for i := range identities {
		if user := userFromIdentity(&identities[i]); user != nil {
			users = append(users, user)
		}
	}
	return users
}

// userFromIdentity returns nil when the identity ID is not a UUID.
func userFromIdentity(identity *ory.Identity) *domain.User {
	id, err := uuid.Parse(identity.Id)
	if err != nil {
		return nil
	}

	return &domain.User{
		ID:        id,
		Email:     traitString(identity.Traits, "email"),
		Username:  traitString(identity.Traits, "username"),
		AvatarURL: traitString(identity.Traits, "avatarUrl"),
		Role:      domain.RoleFromMetadata(identity.MetadataPublic),
		CreatedAt: identity.CreatedAt,
	}
}

func traitString(traits any, key string) string {
	values, ok := traits.(map[string]any)
	if !ok {
		return ""
	}
	value, _ := values[key].(string)
	return value
}
