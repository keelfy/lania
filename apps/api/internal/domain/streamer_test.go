package domain

import (
	"testing"
	"time"
)

func TestStreamerChannelValidate(t *testing.T) {
	tests := []struct {
		name    string
		channel StreamerChannel
		wantErr bool
	}{
		{"twitch", StreamerChannel{StreamerPlatformTwitch, "https://www.twitch.tv/lania"}, false},
		{"youtube handle", StreamerChannel{StreamerPlatformYouTube, "https://youtube.com/@lania"}, false},
		{"youtube short link", StreamerChannel{StreamerPlatformYouTube, "https://youtu.be/abc"}, false},
		{"vk video live", StreamerChannel{StreamerPlatformVKVideo, "https://live.vkvideo.ru/lania"}, false},
		{"kick", StreamerChannel{StreamerPlatformKick, "https://kick.com/lania"}, false},
		{"tiktok upper case host", StreamerChannel{StreamerPlatformTikTok, "https://WWW.TikTok.com/@lania"}, false},
		{"unknown platform", StreamerChannel{"rutube", "https://rutube.ru/lania"}, true},
		{"http", StreamerChannel{StreamerPlatformTwitch, "http://twitch.tv/lania"}, true},
		{"other site", StreamerChannel{StreamerPlatformTwitch, "https://youtube.com/@lania"}, true},
		{"lookalike host", StreamerChannel{StreamerPlatformTwitch, "https://eviltwitch.tv/lania"}, true},
		{"host as path", StreamerChannel{StreamerPlatformTwitch, "https://evil.com/twitch.tv"}, true},
		{"user info", StreamerChannel{StreamerPlatformTwitch, "https://twitch.tv@evil.com/"}, true},
		{"not a link", StreamerChannel{StreamerPlatformTwitch, "lania"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.channel.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStreamerChannels(t *testing.T) {
	twitch := StreamerChannel{StreamerPlatformTwitch, "https://twitch.tv/lania"}
	tests := []struct {
		name     string
		channels []StreamerChannel
		wantErr  bool
	}{
		{"one", []StreamerChannel{twitch}, false},
		{"none", nil, true},
		{"too many", []StreamerChannel{twitch, twitch, twitch, twitch, twitch, twitch}, true},
		{"duplicate with trailing slash", []StreamerChannel{twitch, {StreamerPlatformTwitch, "https://twitch.tv/lania/"}}, true},
		{"one bad", []StreamerChannel{twitch, {StreamerPlatformKick, "https://twitch.tv/other"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateStreamerChannels(tt.channels); (err != nil) != tt.wantErr {
				t.Errorf("ValidateStreamerChannels() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStreamerApplicationReapplyAt(t *testing.T) {
	reviewedAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	rejected := &StreamerApplication{Status: StreamerApplicationStatusRejected, ReviewedAt: &reviewedAt}
	if got := rejected.ReapplyAt(); got == nil || !got.Equal(reviewedAt.Add(StreamerReapplyCooldown)) {
		t.Errorf("ReapplyAt() of a rejected application = %v, want %v", got, reviewedAt.Add(StreamerReapplyCooldown))
	}
	for _, status := range []StreamerApplicationStatus{StreamerApplicationStatusPending, StreamerApplicationStatusApproved} {
		application := &StreamerApplication{Status: status, ReviewedAt: &reviewedAt}
		if got := application.ReapplyAt(); got != nil {
			t.Errorf("ReapplyAt() of a %s application = %v, want nil", status, got)
		}
	}
}
