package requests

import "github.com/google/uuid"

// MarkNotificationsRead names the notifications to stamp. Empty IDs marks every unread notification.
type MarkNotificationsRead struct {
	IDs uuid.UUIDs `json:"ids"`
}

// SendAnnouncement is news for every user. Title and Body map a locale to the text in that language.
type SendAnnouncement struct {
	Title map[string]string `json:"title"`
	Body  map[string]string `json:"body"`
	// Link is a path on the site, like /shop. Empty opens nothing.
	Link string `json:"link"`
}
