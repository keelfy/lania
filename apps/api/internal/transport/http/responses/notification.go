package responses

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID   uuid.UUID `json:"id"`
	Type string    `json:"type"`
	// Payload carries the data of the event. Its shape depends on Type.
	Payload json.RawMessage `json:"payload"`
	// ReadAt is empty while the notification is unread.
	ReadAt    *time.Time `json:"readAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// NotificationList is a page of notifications and the unread counter.
// UnreadCount counts every unread notification, also the ones past the end of Content.
type NotificationList struct {
	Content     []*Notification `json:"content"`
	UnreadCount int64           `json:"unreadCount"`
	// HasMore tells that more notifications follow past the end of Content.
	HasMore bool `json:"hasMore"`
}
