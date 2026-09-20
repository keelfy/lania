package sql

import (
	"context"
	stdsql "database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/lania-smp/backend/internal/domain"
)

type Queries interface {
	// Server Season
	FindSeasonByID(ctx context.Context, seasonID uuid.UUID) (*domain.Season, error)
	// FindSeasons returns every season, the newest number first.
	FindSeasons(ctx context.Context) ([]*domain.Season, error)

	// Game Profile
	GetProfilesByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) ([]*domain.Profile, error)
	FindPublicProfiles(ctx context.Context, search string, only *uuid.UUIDs, sortCol, direction string, size, from int) ([]*domain.Profile, error)
	CountPublicProfiles(ctx context.Context, search string, only *uuid.UUIDs) (int64, error)
	// CountRecentProfiles returns the number of profiles created within the last days.
	CountRecentProfiles(ctx context.Context, days int) (int64, error)
	// FindTopProfilePlaytimes returns players with the most playtime in the season, each with its Profile.
	FindTopProfilePlaytimes(ctx context.Context, seasonID uuid.UUID, limit int) ([]*domain.ProfilePlaytime, error)
	InsertProfile(ctx context.Context, arg InsertProfileParams) error
	ClaimProfile(ctx context.Context, profileID, ownerUserID uuid.UUID, updatedBy uuid.UUID) (bool, error)
	// SetProfileOwner replaces the owner of the profile whoever it is. A nil ownerUserID releases the profile.
	SetProfileOwner(ctx context.Context, profileID uuid.UUID, ownerUserID, updatedBy *uuid.UUID) error
	FindProfileByID(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error)
	FindProfileByMinecraftUUID(ctx context.Context, minecraftUUID uuid.UUID) (*domain.Profile, error)

	// Profile Mojang UUID
	FindProfileMojangUUIDsByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error)
	FindMojangLookupTargets(ctx context.Context, notCheckedSince time.Time, limit int) ([]*domain.MojangLookupTarget, error)
	UpsertProfileMojangUUID(ctx context.Context, mcUUID uuid.UUID, mojangUUID *uuid.UUID) error

	// Profile Cosmetics
	InsertProfileNameColorOption(ctx context.Context, arg InsertProfileNameColorOptionParams) error
	InsertProfileNamePrefixOption(ctx context.Context, arg InsertProfileNamePrefixOptionParams) error
	FindProfileNameColorOptionsByProfileID(ctx context.Context, profileID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error)
	FindProfileNamePrefixOptionsByProfileIDAndType(ctx context.Context, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error)
	FindProfileNameColorOptionByIDAndProfileID(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, seasonID *uuid.UUID) (*domain.ProfileNameColorOption, error)
	FindProfileNamePrefixOptionByIDAndProfileIDAndType(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) (*domain.ProfileNamePrefixOption, error)
	UpdateProfileNameColorByID(ctx context.Context, profileID uuid.UUID, nameColorID uuid.UUID) error
	InsertProfilePrefix(ctx context.Context, arg InsertProfilePrefixParams) error
	FindProfilePrefixesByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfilePrefix, error)
	FindProfilePrefixesByProfileIDs(ctx context.Context, profileIDs uuid.UUIDs) ([]*domain.ProfilePrefix, error)
	FindProfileNameColorOptionsByProfileOwnerUserID(ctx context.Context, ownerUserID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error)
	FindProfileNamePrefixOptionsByProfileOwnerUserIDAndType(ctx context.Context, ownerUserID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error)
	UpdateProfileNamePrefixByProfileIDAndType(ctx context.Context, profileID uuid.UUID, namePrefixID uuid.UUID, prefixType domain.ProfilePrefixType) error
	DeleteProfilePrefixByProfileIDAndType(ctx context.Context, profileID uuid.UUID, prefixType domain.ProfilePrefixType) error

	// Profile Access
	InsertProfileAccess(ctx context.Context, arg InsertProfileAccessParams) error
	CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx context.Context, mcUUID uuid.UUID, seasonID uuid.UUID) (bool, error)
	FindProfileAccessesByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*domain.ProfileAccess, error)
	GetProfileAccessesBySeasonIDAndOwnerUserID(ctx context.Context, seasonID uuid.UUID, ownerUserID uuid.UUID) (uuid.UUIDs, error)

	// Profile Grants
	// FindProfileGrants returns everything the profile was given, revoked grants included, newest first.
	FindProfileGrants(ctx context.Context, profileID, mcUUID uuid.UUID) ([]*domain.Grant, error)
	// The revoke queries leave a grant that is already revoked untouched.
	RevokeProfileAccess(ctx context.Context, mcUUID, accessID uuid.UUID, revokedBy *uuid.UUID) error
	RevokeProfileNameColorOption(ctx context.Context, profileID, optionID uuid.UUID, revokedBy *uuid.UUID) error
	RevokeProfileNamePrefixOption(ctx context.Context, profileID, optionID uuid.UUID, revokedBy *uuid.UUID) error

	// Profile Playtime
	FindProfilePlaytimesByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*domain.ProfilePlaytime, error)
	SumProfilePlaytimesByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]int64, error)
	UpsertProfilePlaytime(ctx context.Context, mcUUID, seasonID uuid.UUID, playtime int64) error
	UpdateProfileSeenAt(ctx context.Context, mcUUID uuid.UUID, firstSeenAt, lastSeenAt *time.Time) error

	// Product
	FindProductsByCategory(ctx context.Context, category domain.ProductCategory, currency domain.Currency, locale string) ([]*domain.Product, error)
	FindProducts(ctx context.Context, locale string, currency domain.Currency) ([]*domain.Product, error)
	FindProductByIDs(ctx context.Context, ids uuid.UUIDs, locale string, currency domain.Currency) ([]*domain.Product, error)

	// Prices
	FindPricesByNames(ctx context.Context, names []domain.ProductPriceName) ([]*domain.ProductPrice, error)

	// Order
	InsertOrder(ctx context.Context, arg InsertOrderParams) (uuid.UUID, error)
	InsertOrderItem(ctx context.Context, arg InsertOrderItemParams) error
	FindOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateOrderStatusByID(ctx context.Context, id uuid.UUID, status domain.OrderStatus, updatedBy *uuid.UUID) error
	FindItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderItem, error)
	FindOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Order, error)
	FindOrderByExternalID(ctx context.Context, externalID string) (*domain.Order, error)
	UpdateOrderExternalIDByID(ctx context.Context, id uuid.UUID, externalID string) error

	// Basket
	InsertBasketItem(ctx context.Context, arg InsertBasketItemParams) error
	FindBasketItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BasketItem, error)
	ClearBasketItemsByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteBasketItemByIDs(ctx context.Context, ids []uuid.UUID) error

	// Integration
	FindOAuth2IntegrationByServiceName(ctx context.Context, serviceName domain.IntegrationService) (*domain.OAuth2Integration, error)
	UpdateOAuth2Integration(ctx context.Context, arg UpdateOAuth2IntegrationParams) error

	// Easy Donate
	FindEDProductsByProductIDs(ctx context.Context, productIDs uuid.UUIDs) ([]int64, error)
}

type queryable interface {
	ExecContext(ctx context.Context, query string, args ...any) (stdsql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*stdsql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *stdsql.Row
}

type queries struct {
	x queryable
}

// NewMySQL returns Queries backed by database/sql MySQL driver
func NewMySQL(db *stdsql.DB) Queries {
	return &queries{x: db}
}

// WithMySQLTx returns Queries bound to a single transaction
func WithMySQLTx(tx *stdsql.Tx) Queries {
	return &queries{x: tx}
}
