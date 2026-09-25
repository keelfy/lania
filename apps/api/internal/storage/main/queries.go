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
	FindPrimarySeasonID(ctx context.Context) (uuid.UUID, error)
	FindPublicSeasons(ctx context.Context) ([]*domain.Season, error)
	// FindSeasons returns every season, the newest start first.
	FindSeasons(ctx context.Context) ([]*domain.Season, error)
	InsertSeason(ctx context.Context, arg InsertSeasonParams) error
	UpdateSeason(ctx context.Context, arg UpdateSeasonParams) error
	DeleteSeason(ctx context.Context, seasonID uuid.UUID) (bool, error)
	ClearPrimarySeasons(ctx context.Context) error
	SetSeasonPrimary(ctx context.Context, seasonID uuid.UUID) (bool, error)
	SetSeasonShellAddress(ctx context.Context, seasonID uuid.UUID, address string) error

	// Season Screenshot
	// FindSeasonScreenshots returns every screenshot of the season, in feed order, each with its authors in credit order.
	FindSeasonScreenshots(ctx context.Context, seasonID uuid.UUID) ([]*domain.SeasonScreenshot, error)
	// FindSeasonScreenshotByID returns stdsql.ErrNoRows when the screenshot does not exist in the season.
	FindSeasonScreenshotByID(ctx context.Context, screenshotID, seasonID uuid.UUID) (*domain.SeasonScreenshot, error)
	CreateSeasonScreenshot(ctx context.Context, arg SaveSeasonScreenshotParams) error
	// UpdateSeasonScreenshot reports whether a screenshot with the id existed in the season.
	UpdateSeasonScreenshot(ctx context.Context, arg SaveSeasonScreenshotParams) (bool, error)
	// DeleteSeasonScreenshot reports whether the screenshot existed in the season.
	DeleteSeasonScreenshot(ctx context.Context, screenshotID, seasonID uuid.UUID) (bool, error)
	// SetSeasonScreenshotAuthors replaces every credit of the screenshot with profileIDs, in the given order.
	SetSeasonScreenshotAuthors(ctx context.Context, screenshotID uuid.UUID, profileIDs uuid.UUIDs) error

	// Season World
	// FindSeasonWorlds returns the worlds of the season in display order.
	FindSeasonWorlds(ctx context.Context, seasonID uuid.UUID) ([]*domain.SeasonWorld, error)
	// FindSeasonWorldByID returns stdsql.ErrNoRows when the world does not exist.
	FindSeasonWorldByID(ctx context.Context, worldID uuid.UUID) (*domain.SeasonWorld, error)
	CreateSeasonWorld(ctx context.Context, arg SaveSeasonWorldParams) error
	// UpdateSeasonWorld reports whether a world with the id existed.
	UpdateSeasonWorld(ctx context.Context, arg SaveSeasonWorldParams) (bool, error)
	// DeleteSeasonWorld reports whether the world existed.
	DeleteSeasonWorld(ctx context.Context, worldID uuid.UUID) (bool, error)

	// Chunk Claim
	// FindActiveChunkClaims returns every claim that holds in the dimension of the world, each with its profile.
	FindActiveChunkClaims(ctx context.Context, worldID uuid.UUID, dimension string) ([]*domain.ChunkClaim, error)
	// CountActiveChunkClaimsByProfile counts the claims the profile holds in the world, over every dimension.
	CountActiveChunkClaimsByProfile(ctx context.Context, profileID, worldID uuid.UUID) (int, error)
	// CountActiveChunkClaimsAt returns how many of the chunks are already claimed by anyone.
	CountActiveChunkClaimsAt(ctx context.Context, worldID uuid.UUID, dimension string, chunks []domain.ChunkPos) (int, error)
	InsertChunkClaims(ctx context.Context, worldID uuid.UUID, dimension string, profileID uuid.UUID, chunks []domain.ChunkPos) error
	// ReleaseChunkClaims returns how many claims it ended; a non-nil profileID limits it to that profile's claims.
	ReleaseChunkClaims(ctx context.Context, worldID uuid.UUID, dimension string, profileID *uuid.UUID, chunks []domain.ChunkPos, releasedBy uuid.UUID) (int64, error)
	// CountChunkClaimsInWorld counts every claim the world ever had, released ones too.
	CountChunkClaimsInWorld(ctx context.Context, worldID uuid.UUID) (int, error)
	// LockProfile holds the profile row until the transaction ends.
	LockProfile(ctx context.Context, profileID uuid.UUID) error

	// Game Profile
	GetProfilesByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) ([]*domain.Profile, error)
	// FindPublicProfiles sorts by last_seen_at in seasonID, or over every season when seasonID is uuid.Nil.
	FindPublicProfiles(ctx context.Context, search string, only *uuid.UUIDs, sortCol, direction string, seasonID uuid.UUID, size, from int) ([]*domain.Profile, error)
	CountPublicProfiles(ctx context.Context, search string, only *uuid.UUIDs) (int64, error)
	// CountRecentProfiles returns the number of profiles created within the last days.
	CountRecentProfiles(ctx context.Context, days int) (int64, error)
	// FindTopProfilePlaytimes returns players with the most playtime in the season, each with its Profile.
	FindTopProfilePlaytimes(ctx context.Context, seasonID uuid.UUID, limit int) ([]*domain.ProfilePlaytime, error)
	InsertProfile(ctx context.Context, arg InsertProfileParams) error
	ClaimProfile(ctx context.Context, profileID, ownerUserID uuid.UUID, updatedBy uuid.UUID) (bool, error)
	// SetProfileOwner replaces the owner of the profile whoever it is. A nil ownerUserID releases the profile.
	SetProfileOwner(ctx context.Context, profileID uuid.UUID, ownerUserID, updatedBy *uuid.UUID) error
	// SetProfileRole stores the role and marks the moment of the change for FindProfileRolesChangedSince.
	SetProfileRole(ctx context.Context, profileID uuid.UUID, role domain.Role, updatedBy *uuid.UUID) error
	// FindProfileRolesChangedSince returns the role of every profile whose role changed within the last since.
	FindProfileRolesChangedSince(ctx context.Context, since time.Duration) ([]*domain.ProfileRole, error)
	// FindMinecraftUUIDsByRoles returns the players whose stored role is one of the roles.
	FindMinecraftUUIDsByRoles(ctx context.Context, roles []domain.Role) (uuid.UUIDs, error)
	FindProfileByID(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error)
	FindProfileByMinecraftUUID(ctx context.Context, minecraftUUID uuid.UUID) (*domain.Profile, error)
	// FindProfileByUsername matches the username without regard to case, like Minecraft does.
	FindProfileByUsername(ctx context.Context, username string) (*domain.Profile, error)
	// RekeyProfile moves the profile from oldMcUUID to newMcUUID and keeps the first UUID it ever had as legacy.
	// It reports false when the profile is no longer at oldMcUUID.
	RekeyProfile(ctx context.Context, profileID, oldMcUUID, newMcUUID uuid.UUID) (bool, error)
	SetProfilePremiumConflict(ctx context.Context, mcUUID uuid.UUID) error
	// FindLegacyMinecraftUUIDs maps the legacy UUID of every rekeyed profile to its current one, for profiles where
	// either of the two is among mcUUIDs.
	FindLegacyMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error)
	// FindProfilesByIDs returns the profiles that exist among ids; a missing id is simply absent from the result.
	FindProfilesByIDs(ctx context.Context, ids uuid.UUIDs) ([]*domain.Profile, error)

	// Profile Mojang UUID
	FindProfileMojangUUIDsByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error)
	FindMojangLookupTargets(ctx context.Context, notCheckedSince time.Time, limit int) ([]*domain.MojangLookupTarget, error)
	UpsertProfileMojangUUID(ctx context.Context, mcUUID uuid.UUID, mojangUUID *uuid.UUID) error
	// FindPremiumRekeyTargets returns profiles whose nickname has a Mojang account while mc_uuid is not that account,
	// premium conflicts excluded.
	FindPremiumRekeyTargets(ctx context.Context) ([]*domain.PremiumRekeyTarget, error)
	// CountUncheckedMojangProfiles counts profiles whose nickname was never looked up on Mojang.
	CountUncheckedMojangProfiles(ctx context.Context) (int, error)

	// Profile Verification
	// UpsertProfileVerification opens a new verification request of the profile for ttl, replacing an earlier one.
	UpsertProfileVerification(ctx context.Context, profileID uuid.UUID, ttl time.Duration) error
	// FindOpenProfileVerification returns stdsql.ErrNoRows when the profile has no unexpired request.
	FindOpenProfileVerification(ctx context.Context, profileID uuid.UUID) (*domain.ProfileVerification, error)
	// SetProfileVerificationCode stores the issued code and keeps the request open for ttl more.
	SetProfileVerificationCode(ctx context.Context, profileID uuid.UUID, code string, ttl time.Duration) error
	IncrementProfileVerificationAttempts(ctx context.Context, profileID uuid.UUID) error
	DeleteProfileVerification(ctx context.Context, profileID uuid.UUID) error
	// SetProfileVerified marks mcUUID as the licensed account the owner proved.
	SetProfileVerified(ctx context.Context, profileID, mcUUID uuid.UUID) error

	// Profile Cosmetics
	// FindNameColors returns every name color ordered by name.
	FindNameColors(ctx context.Context) ([]*domain.NameColor, error)
	// FindNamePrefixes returns every name prefix ordered by name.
	FindNamePrefixes(ctx context.Context) ([]*domain.NamePrefix, error)
	FindNameColorByID(ctx context.Context, nameColorID uuid.UUID) (*domain.NameColor, error)
	FindNamePrefixByID(ctx context.Context, namePrefixID uuid.UUID) (*domain.NamePrefix, error)
	InsertNameColor(ctx context.Context, id uuid.UUID, name string, colors []string) error
	UpdateNameColor(ctx context.Context, id uuid.UUID, name string, colors []string) error
	InsertNamePrefix(ctx context.Context, id uuid.UUID, name string, metadata domain.NamePrefixMetadata) error
	UpdateNamePrefix(ctx context.Context, id uuid.UUID, name string, metadata domain.NamePrefixMetadata) error
	InsertProfileNameColorOption(ctx context.Context, arg InsertProfileNameColorOptionParams) error
	InsertProfileNamePrefixOption(ctx context.Context, arg InsertProfileNamePrefixOptionParams) error
	FindProfileNameColorOptionsByProfileID(ctx context.Context, profileID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error)
	FindProfileNamePrefixOptionsByProfileIDAndType(ctx context.Context, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error)
	FindProfileNameColorOptionByIDAndProfileID(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, seasonID *uuid.UUID) (*domain.ProfileNameColorOption, error)
	FindProfileNamePrefixOptionByIDAndProfileIDAndType(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) (*domain.ProfileNamePrefixOption, error)
	// FindProfilesSeasonCosmetics returns what every profile shows in the season, keyed by profile ID.
	// A profile that picked no name color gets defaultNameColorID.
	FindProfilesSeasonCosmetics(ctx context.Context, profileIDs uuid.UUIDs, seasonID, defaultNameColorID uuid.UUID) (map[uuid.UUID]*domain.ProfileCosmetics, error)
	SetProfileSeasonNameColor(ctx context.Context, profileID, seasonID, nameColorID uuid.UUID) error
	// SetProfileSeasonPrefix selects the name prefix of the type. A nil namePrefixID clears it.
	SetProfileSeasonPrefix(ctx context.Context, profileID, seasonID uuid.UUID, prefixType domain.ProfilePrefixType, namePrefixID *uuid.UUID) error
	// PruneProfileSeasonCosmetics resets every selection of the profile in every season that no unrevoked option covers.
	PruneProfileSeasonCosmetics(ctx context.Context, profileID uuid.UUID) error
	FindProfileNameColorOptionsByProfileOwnerUserID(ctx context.Context, ownerUserID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error)
	FindProfileNamePrefixOptionsByProfileOwnerUserIDAndType(ctx context.Context, ownerUserID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error)

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
	// RevokeProfileCosmetics revokes every name color and name prefix the profile still has, except the name color keepNameColorID.
	RevokeProfileCosmetics(ctx context.Context, profileID, keepNameColorID uuid.UUID, revokedBy *uuid.UUID) error

	// Profile Playtime
	FindProfilePlaytimesByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*domain.ProfilePlaytime, error)
	SumProfilePlaytimesByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]int64, error)
	// FindProfileSeasonStats returns the stats of the profile in every season it played in, the newest season first.
	FindProfileSeasonStats(ctx context.Context, mcUUID uuid.UUID) ([]*domain.ProfileSeasonStats, error)
	// UpsertProfilePlaytime stores the playtime and moves the last seen date of the profile in the season forward.
	// A nil lastSeenAt keeps the stored date.
	UpsertProfilePlaytime(ctx context.Context, mcUUID, seasonID uuid.UUID, playtime int64, lastSeenAt *time.Time) error
	// FindProfilesLastSeenInSeason returns when every profile was last seen in the season.
	// Profiles with no known date are missing from the result.
	FindProfilesLastSeenInSeason(ctx context.Context, mcUUIDs uuid.UUIDs, seasonID uuid.UUID) (map[uuid.UUID]time.Time, error)
	UpdateProfileSeenAt(ctx context.Context, mcUUID uuid.UUID, firstSeenAt, lastSeenAt *time.Time) error

	// Product
	FindProductsByCategory(ctx context.Context, category domain.ProductCategory, currency domain.Currency, locale string) ([]*domain.Product, error)
	FindProducts(ctx context.Context, locale string, currency domain.Currency) ([]*domain.Product, error)
	FindProductByIDs(ctx context.Context, ids uuid.UUIDs, locale string, currency domain.Currency) ([]*domain.Product, error)
	FindProductByIDsIncludingInactive(ctx context.Context, ids uuid.UUIDs, locale string, currency domain.Currency) ([]*domain.Product, error)
	FindAdminProducts(ctx context.Context) ([]*domain.Product, error)
	InsertProduct(ctx context.Context, arg SaveProductParams) error
	UpdateProduct(ctx context.Context, arg SaveProductParams) error
	UpsertProductLocalization(ctx context.Context, productID uuid.UUID, locale, name, description string) error
	UpsertEDProduct(ctx context.Context, productID uuid.UUID, edProductID int64) error
	DeleteEDProduct(ctx context.Context, productID uuid.UUID) error
	DeleteProduct(ctx context.Context, productID uuid.UUID) error

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
	FindAdminOrders(ctx context.Context, filter domain.OrderFilter, size, from int) ([]*domain.Order, error)
	CountAdminOrders(ctx context.Context, filter domain.OrderFilter) (int64, error)
	CountOrdersByStatus(ctx context.Context) (map[domain.OrderStatus]int64, error)
	FindAdminOrderItems(ctx context.Context, orderIDs uuid.UUIDs, locale string) ([]*domain.OrderItem, error)

	// Basket
	InsertBasketItem(ctx context.Context, arg InsertBasketItemParams) error
	FindBasketItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BasketItem, error)
	ClearBasketItemsByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteBasketItemByIDs(ctx context.Context, ids []uuid.UUID) error

	// Easy Donate
	FindEDProductsByProductIDs(ctx context.Context, productIDs uuid.UUIDs) ([]int64, error)

	// Profile Merge
	// LockProfilesForMerge locks both profile rows for the duration of the transaction.
	LockProfilesForMerge(ctx context.Context, firstProfileID, secondProfileID uuid.UUID) error
	// FindLiveSyncedSeasonNamesWithPlaytime returns the seasons where the profile has playtime and a shell
	// still syncs it every minute.
	FindLiveSyncedSeasonNamesWithPlaytime(ctx context.Context, mcUUID uuid.UUID) ([]string, error)
	// MergeProfileData moves every table that references the source profile into the target profile.
	MergeProfileData(ctx context.Context, sourceProfileID, sourceMcUUID, targetProfileID, targetMcUUID uuid.UUID) (*domain.ProfileMergeCounts, error)
	// UpdateProfileAfterMerge writes the merged owner and seen dates onto the target profile.
	UpdateProfileAfterMerge(ctx context.Context, targetProfileID uuid.UUID, ownerUserID *uuid.UUID, firstSeenAt, lastSeenAt *time.Time, updatedBy uuid.UUID) error
	// DeleteProfile removes the profile row. Every table that referenced it must be cleared first.
	DeleteProfile(ctx context.Context, profileID uuid.UUID) error
	InsertProfileMerge(ctx context.Context, arg InsertProfileMergeParams) error
	// FindProfileMergesByTargetProfileID returns every profile merged into the profile, newest first.
	FindProfileMergesByTargetProfileID(ctx context.Context, targetProfileID uuid.UUID) ([]*domain.ProfileMerge, error)

	// Notification
	InsertNotification(ctx context.Context, arg InsertNotificationParams) error
	// FindNotificationsByUserID returns the notifications of the user that match the filter, newest first.
	FindNotificationsByUserID(ctx context.Context, userID uuid.UUID, filter domain.NotificationFilter) ([]*domain.Notification, error)
	CountUnreadNotificationsByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkNotificationsRead stamps the unread notifications of the user. Empty ids marks all of them.
	MarkNotificationsRead(ctx context.Context, userID uuid.UUID, ids uuid.UUIDs) error
	DeleteNotificationsByUserID(ctx context.Context, userID uuid.UUID) error
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
