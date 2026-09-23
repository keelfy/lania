package responses

import "github.com/google/uuid"

type AdminUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username,omitempty"`
	AvatarURL string    `json:"avatarUrl,omitempty"`
	Role      string    `json:"role"`
	CreatedAt *int64    `json:"createdAt,omitempty"`
}

type AdminUserDetails struct {
	AdminUser
	Profiles []*AdminProfile `json:"profiles"`
}

type AdminProfile struct {
	ID            uuid.UUID  `json:"id"`
	MinecraftUUID uuid.UUID  `json:"mcUuid"`
	Username      string     `json:"username"`
	OwnerUserID   *uuid.UUID `json:"ownerUserId,omitempty"`
	Role          string     `json:"role"`
	CreatedAt     int64      `json:"createdAt"`
}

type AdminProfileDetails struct {
	AdminProfile
	// Owner is empty while nobody owns the profile.
	Owner *AdminUser `json:"owner,omitempty"`
}

type ProfileMergeBlocker struct {
	Kind        string   `json:"kind"`
	SeasonNames []string `json:"seasonNames,omitempty"`
}

type ProfileMergeCounts struct {
	PlaytimeMoved            int64 `json:"playtimeMoved"`
	PlaytimeSummed           int64 `json:"playtimeSummed"`
	AccessesMoved            int64 `json:"accessesMoved"`
	AccessesDropped          int64 `json:"accessesDropped"`
	ViolationsMoved          int64 `json:"violationsMoved"`
	NameColorOptionsMoved    int64 `json:"nameColorOptionsMoved"`
	NameColorOptionsDropped  int64 `json:"nameColorOptionsDropped"`
	NamePrefixOptionsMoved   int64 `json:"namePrefixOptionsMoved"`
	NamePrefixOptionsDropped int64 `json:"namePrefixOptionsDropped"`
	SeasonCosmeticsMoved     int64 `json:"seasonCosmeticsMoved"`
	SeasonCosmeticsDropped   int64 `json:"seasonCosmeticsDropped"`
	PrefixesMoved            int64 `json:"prefixesMoved"`
	PrefixesDropped          int64 `json:"prefixesDropped"`
	OrderItemsMoved          int64 `json:"orderItemsMoved"`
	BasketItemsMoved         int64 `json:"basketItemsMoved"`
	BasketItemsDropped       int64 `json:"basketItemsDropped"`
	NotificationsRepointed   int64 `json:"notificationsRepointed"`
}

type ProfileMergeSummary struct {
	SourceProfileID uuid.UUID              `json:"sourceProfileId"`
	SourceUsername  string                 `json:"sourceUsername"`
	TargetProfileID uuid.UUID              `json:"targetProfileId"`
	TargetUsername  string                 `json:"targetUsername"`
	RoleBefore      string                 `json:"roleBefore"`
	RoleAfter       string                 `json:"roleAfter"`
	OwnerUserID     *uuid.UUID             `json:"ownerUserId,omitempty"`
	Counts          ProfileMergeCounts     `json:"counts"`
	Blockers        []*ProfileMergeBlocker `json:"blockers,omitempty"`
	CanMerge        bool                   `json:"canMerge"`
	// Resync is the report of updating the season servers. Empty for a preview or when the merge failed before it.
	Resync *ProfileResync `json:"resync,omitempty"`
}

type ProfileMerge struct {
	ID                  uuid.UUID `json:"id"`
	SourceProfileID     uuid.UUID `json:"sourceProfileId"`
	SourceMinecraftUUID uuid.UUID `json:"sourceMcUuid"`
	SourceUsername      string    `json:"sourceUsername"`
	MergedBy            uuid.UUID `json:"mergedBy"`
	CreatedAt           int64     `json:"createdAt"`
}

type AdminGrant struct {
	ID          uuid.UUID  `json:"id"`
	Type        string     `json:"type"`
	SeasonID    *uuid.UUID `json:"seasonId,omitempty"`
	Name        string     `json:"name,omitempty"`
	PrefixType  string     `json:"prefixType,omitempty"`
	Source      string     `json:"source,omitempty"`
	OrderItemID *uuid.UUID `json:"orderItemId,omitempty"`
	GrantedBy   *uuid.UUID `json:"grantedBy,omitempty"`
	CreatedAt   int64      `json:"createdAt"`
	RevokedAt   *int64     `json:"revokedAt,omitempty"`
	RevokedBy   *uuid.UUID `json:"revokedBy,omitempty"`
}

type AdminNameColor struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Colors []string  `json:"colors"`
}

type AdminNamePrefix struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Image   string    `json:"image"`
	Prefix  string    `json:"prefix"`
	NoSpace bool      `json:"noSpace"`
}

type AdminCosmeticsCatalog struct {
	NameColors   []*AdminNameColor  `json:"nameColors"`
	NamePrefixes []*AdminNamePrefix `json:"namePrefixes"`
}

type AdminProductLocalization struct {
	Locale      string `json:"locale"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AdminProductPrice struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type AdminProduct struct {
	ID                  uuid.UUID                   `json:"id"`
	Category            string                      `json:"category"`
	PriceName           string                      `json:"priceName"`
	Metadata            any                         `json:"metadata"`
	IsActive            bool                        `json:"isActive"`
	EasyDonateProductID *int64                      `json:"easyDonateProductId,omitempty"`
	SoldCount           int64                       `json:"soldCount"`
	Localizations       []*AdminProductLocalization `json:"localizations"`
	Prices              []*AdminProductPrice        `json:"prices"`
}

// UploadedImage is the location of an image an admin just uploaded to the project's S3 bucket.
type UploadedImage struct {
	Location string `json:"location"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}
