package requests

import "github.com/google/uuid"

// MarkNotificationsRead names the notifications to stamp. Empty IDs marks every unread notification.
type MarkNotificationsRead struct {
	IDs uuid.UUIDs `json:"ids"`
}
