package responses

import "github.com/google/uuid"

type ScreenshotAuthor struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type SeasonScreenshot struct {
	ID       uuid.UUID          `json:"id"`
	Image    string             `json:"image"`
	Title    *string            `json:"title,omitempty"`
	Position int                `json:"position"`
	Authors  []ScreenshotAuthor `json:"authors"`
}
