package domain

import (
	"time"

	"github.com/google/uuid"
)

// SeasonScreenshot is one photo of a season, shown in the feed on the season page.
type SeasonScreenshot struct {
	ID       uuid.UUID
	SeasonID uuid.UUID
	// Image is an s3://bucket/key location, the same form as Season.PreviewImage.
	Image     string
	Title     *string
	Position  int
	CreatedAt time.Time
	// Authors is populated by the query that loads the screenshot, not written directly.
	Authors []*Profile
}
