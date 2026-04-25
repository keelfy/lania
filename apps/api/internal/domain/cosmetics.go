package domain

import "github.com/google/uuid"

type NameColorMetadata struct {
	Colors []string `json:"colors"`
}

type NamePrefixMetadata struct {
	Prefix  string `json:"prefix"`
	Image   string `json:"image"`
	NoSpace bool   `json:"noSpace"`
}

type NameColor struct {
	ID       uuid.UUID
	Name     string
	Metadata NameColorMetadata
}

type NamePrefix struct {
	ID       uuid.UUID
	Name     string
	Metadata NamePrefixMetadata
}
