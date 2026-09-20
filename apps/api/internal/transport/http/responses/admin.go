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
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Image string    `json:"image"`
}

type AdminCosmeticsCatalog struct {
	NameColors   []*AdminNameColor  `json:"nameColors"`
	NamePrefixes []*AdminNamePrefix `json:"namePrefixes"`
}
