package commands

import (
	"strings"
	"testing"
)

func TestSendAnnouncementCommand_Validate(t *testing.T) {
	t.Parallel()

	title := map[string]string{"ru": "Новые товары", "en": "New products"}
	body := map[string]string{"ru": "Загляните в магазин", "en": "Check out the shop"}

	tests := []struct {
		name    string
		command SendAnnouncementCommand
		wantErr bool
	}{
		{"title only", SendAnnouncementCommand{Title: title}, false},
		{"title, body and link", SendAnnouncementCommand{Title: title, Body: body, Link: "/shop?tab=cosmetics"}, false},
		{"blank bodies", SendAnnouncementCommand{Title: title, Body: map[string]string{"ru": "", "en": ""}}, false},
		{"root link", SendAnnouncementCommand{Title: title, Link: "/"}, false},
		{"no title", SendAnnouncementCommand{}, true},
		{"title missing a locale", SendAnnouncementCommand{Title: map[string]string{"ru": "Новые товары"}}, true},
		{"unsupported locale", SendAnnouncementCommand{Title: map[string]string{"ru": "a", "en": "b", "de": "c"}}, true},
		{"title too long", SendAnnouncementCommand{Title: map[string]string{"ru": strings.Repeat("а", 121), "en": "b"}}, true},
		{"body missing a locale", SendAnnouncementCommand{Title: title, Body: map[string]string{"en": "Check out the shop"}}, true},
		{"absolute link", SendAnnouncementCommand{Title: title, Link: "https://evil.example"}, true},
		{"protocol-relative link", SendAnnouncementCommand{Title: title, Link: "//evil.example"}, true},
		{"backslash link", SendAnnouncementCommand{Title: title, Link: `/\evil.example`}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.command.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendAnnouncementCommand_Payload(t *testing.T) {
	t.Parallel()

	cmd := SendAnnouncementCommand{
		Title: map[string]string{"ru": "а", "en": "b"},
		Body:  map[string]string{"ru": "", "en": ""},
	}
	if payload := cmd.Payload(); payload.Body != nil {
		t.Errorf("blank bodies must be left out, got %v", payload.Body)
	}
}
