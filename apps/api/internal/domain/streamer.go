package domain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StreamerGroup is the group of the streamer role on the season servers. It gives the stream announcement command.
const StreamerGroup = "streamer"

const (
	MinStreamerChannels          = 1
	MaxStreamerChannels          = 5
	MaxStreamerDescriptionLength = 500
	MaxStreamerAboutLength       = 2000
	MaxStreamerRejectReason      = 500
	// StreamerReapplyCooldown is how long an owner waits to apply again after a rejection.
	StreamerReapplyCooldown = 7 * 24 * time.Hour
)

type StreamerPlatform string

const (
	StreamerPlatformTwitch  StreamerPlatform = "twitch"
	StreamerPlatformYouTube StreamerPlatform = "youtube"
	StreamerPlatformVKVideo StreamerPlatform = "vk_video"
	StreamerPlatformKick    StreamerPlatform = "kick"
	StreamerPlatformTikTok  StreamerPlatform = "tiktok"
)

// streamerPlatformDomains are the sites a channel of the platform lives on. Their subdomains count too.
var streamerPlatformDomains = map[StreamerPlatform][]string{
	StreamerPlatformTwitch:  {"twitch.tv"},
	StreamerPlatformYouTube: {"youtube.com", "youtu.be"},
	StreamerPlatformVKVideo: {"vkvideo.ru", "vk.com", "vk.ru"},
	StreamerPlatformKick:    {"kick.com"},
	StreamerPlatformTikTok:  {"tiktok.com"},
}

// StreamerChannel is a channel of a streamer on a streaming platform.
type StreamerChannel struct {
	Platform StreamerPlatform `json:"platform"`
	URL      string           `json:"url"`
}

// Validate checks that the URL is an https link to a channel on the site of the platform.
func (c StreamerChannel) Validate() error {
	domains, ok := streamerPlatformDomains[c.Platform]
	if !ok {
		return fmt.Errorf("unknown platform %q", c.Platform)
	}
	u, err := url.Parse(c.URL)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return fmt.Errorf("channel link %q is not an https link", c.URL)
	}
	host := strings.ToLower(u.Hostname())
	for _, domain := range domains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return nil
		}
	}
	return fmt.Errorf("channel link %q is not on %s", c.URL, c.Platform)
}

// ValidateStreamerChannels checks the number of channels and every channel.
func ValidateStreamerChannels(channels []StreamerChannel) error {
	if len(channels) < MinStreamerChannels || len(channels) > MaxStreamerChannels {
		return fmt.Errorf("a streamer has %d to %d channels", MinStreamerChannels, MaxStreamerChannels)
	}
	seen := make(map[string]bool, len(channels))
	var errs []error
	for _, channel := range channels {
		if err := channel.Validate(); err != nil {
			errs = append(errs, err)
			continue
		}
		key := strings.ToLower(strings.TrimRight(channel.URL, "/"))
		if seen[key] {
			errs = append(errs, fmt.Errorf("channel link %q is listed twice", channel.URL))
		}
		seen[key] = true
	}
	return errors.Join(errs...)
}

// Streamer is a profile with the streamer role.
type Streamer struct {
	ProfileID   uuid.UUID
	Channels    []StreamerChannel
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// Profile is filled in lists that show the streamer.
	Profile *Profile
}

type StreamerApplicationStatus string

const (
	StreamerApplicationStatusPending  StreamerApplicationStatus = "pending"
	StreamerApplicationStatusApproved StreamerApplicationStatus = "approved"
	StreamerApplicationStatusRejected StreamerApplicationStatus = "rejected"
)

func (s StreamerApplicationStatus) Valid() bool {
	switch s {
	case StreamerApplicationStatusPending, StreamerApplicationStatusApproved, StreamerApplicationStatusRejected:
		return true
	}
	return false
}

// StreamerApplication is a request of an owner for the streamer role of the profile.
type StreamerApplication struct {
	ID           uuid.UUID
	ProfileID    uuid.UUID
	UserID       uuid.UUID
	Channels     []StreamerChannel
	About        string
	Status       StreamerApplicationStatus
	RejectReason *string
	ReviewedBy   *uuid.UUID
	ReviewedAt   *time.Time
	CreatedAt    time.Time
	// Profile is filled in the admin queue.
	Profile *Profile
}

// ReapplyAt is when the owner can apply again after the application was rejected, nil for other statuses.
func (a *StreamerApplication) ReapplyAt() *time.Time {
	if a.Status != StreamerApplicationStatusRejected || a.ReviewedAt == nil {
		return nil
	}
	at := a.ReviewedAt.Add(StreamerReapplyCooldown)
	return &at
}
