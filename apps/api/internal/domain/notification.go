package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	// NotificationTypeCosmeticGranted tells the user that a profile of theirs was given a name color or a name prefix.
	NotificationTypeCosmeticGranted NotificationType = "cosmetic-granted"
	// NotificationTypeCosmeticRevoked tells the user that a name color or a name prefix was taken back.
	NotificationTypeCosmeticRevoked NotificationType = "cosmetic-revoked"
	// NotificationTypeProfileMerged tells the owner that an admin moved another profile's data into this one.
	NotificationTypeProfileMerged NotificationType = "profile-merged"
)

func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeCosmeticGranted, NotificationTypeCosmeticRevoked, NotificationTypeProfileMerged:
		return true
	}
	return false
}

// Notification is one entry of the bell menu. The text is built by the frontend from Type and Payload,
// so the same notification reads in the language the user picked.
type Notification struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Type    NotificationType
	Payload json.RawMessage
	ReadAt  *time.Time
	// CreatedAt is when the event happened.
	CreatedAt time.Time
}

func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}

// CosmeticNotificationPayload is the payload of NotificationTypeCosmeticGranted and NotificationTypeCosmeticRevoked.
type CosmeticNotificationPayload struct {
	ProfileID uuid.UUID `json:"profileId"`
	// ProfileUsername is copied in so an old notification keeps the name it was made with.
	ProfileUsername string `json:"profileUsername"`
	// GrantType is name-color or name-prefix.
	GrantType GrantType `json:"grantType"`
	// PrefixType is set for a name prefix only.
	PrefixType ProfilePrefixType `json:"prefixType,omitempty"`
	ItemID     uuid.UUID         `json:"itemId"`
	// ItemName is copied in so the notification survives a rename of the item.
	ItemName string `json:"itemName"`
	// Colors is set for a name color only, so the bell menu can paint the name without asking for the item.
	Colors []string `json:"colors,omitempty"`
	// PrefixImage is set for a name prefix only.
	PrefixImage string `json:"prefixImage,omitempty"`
	// SeasonID is empty for a cosmetic that is granted for good.
	SeasonID *uuid.UUID `json:"seasonId,omitempty"`
	// SeasonName is copied in next to SeasonID, so the bell menu names the season without asking for it.
	SeasonName string `json:"seasonName,omitempty"`
}

// ProfileMergeNotificationPayload is the payload of NotificationTypeProfileMerged.
type ProfileMergeNotificationPayload struct {
	// ProfileID is the target profile, the one that received the data.
	ProfileID uuid.UUID `json:"profileId"`
	// ProfileUsername is copied in so the notification keeps the name it was made with.
	ProfileUsername string `json:"profileUsername"`
	// SourceUsername is the nickname the data was carried over from. The source profile no longer exists.
	SourceUsername string `json:"sourceUsername"`
}

// NotificationFilter narrows the notifications of a user.
type NotificationFilter struct {
	// UnreadOnly leaves out the notifications that were read.
	UnreadOnly bool
	Offset     int
	Limit      int
}
