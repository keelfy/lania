package domain

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

// ProfileCursor is the position of a profile in a list. Profiles are ordered by a sort value, then by id,
// so the two together are a unique key. The value is kept as the database wrote it, and is nil for a profile
// that has none.
type ProfileCursor struct {
	Value *string   `json:"v"`
	ID    uuid.UUID `json:"i"`
}

// Encode returns the opaque token that clients send back to get the next page.
func (c *ProfileCursor) Encode() string {
	payload, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(payload)
}

func DecodeProfileCursor(token string) (*ProfileCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var cursor ProfileCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return nil, err
	}
	if cursor.ID == uuid.Nil {
		return nil, errors.New("incomplete profile cursor")
	}
	return &cursor, nil
}

// ProfilePage is a page of profiles. Total is the number of profiles matching the filter, and is only
// counted for the first page.
type ProfilePage struct {
	Profiles   []*Profile
	NextCursor *ProfileCursor
	Total      int64
}
