package commands

import (
	"fmt"
	"regexp"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/lania-smp/backend/internal/domain"
)

const (
	maxAnnouncementTitleLength = 120
	maxAnnouncementBodyLength  = 1000
	maxAnnouncementLinkLength  = 255
)

// sitePathPattern accepts a path on the site like /shop or /seasons/1?tab=worlds.
// It rejects //host and /\host, which a browser opens as another site.
var sitePathPattern = regexp.MustCompile(`^/([^/\\\s]\S*)?$`)

// SendAnnouncementCommand is news an admin sends to every user. Title and Body map a locale to the text.
type SendAnnouncementCommand struct {
	Title map[string]string
	Body  map[string]string
	Link  string
}

func (c *SendAnnouncementCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Title, validation.By(validateLocalizedText(c.Title, true, maxAnnouncementTitleLength))),
		// A body is optional, but once written it is written in every language, so no reader gets half the news.
		validation.Field(&c.Body, validation.By(validateLocalizedText(c.Body, hasText(c.Body), maxAnnouncementBodyLength))),
		validation.Field(&c.Link,
			validation.RuneLength(0, maxAnnouncementLinkLength),
			validation.Match(sitePathPattern).Error("must be a path on the site starting with /"),
		),
	)
}

// Payload is what every user gets stored, the texts without the empty ones.
func (c *SendAnnouncementCommand) Payload() domain.AnnouncementNotificationPayload {
	payload := domain.AnnouncementNotificationPayload{Title: c.Title, Link: c.Link}
	if hasText(c.Body) {
		payload.Body = c.Body
	}
	return payload
}

// validateLocalizedText checks that texts holds the supported locales only and, when required,
// a text of 1 to maxLength runes in each of them.
func validateLocalizedText(texts map[string]string, required bool, maxLength int) validation.RuleFunc {
	return func(any) error {
		for locale, text := range texts {
			if !slices.Contains(domain.SupportedLocales, locale) {
				return fmt.Errorf("locale %q is not supported", locale)
			}
			if len([]rune(text)) > maxLength {
				return fmt.Errorf("%s: must be at most %d characters", locale, maxLength)
			}
		}
		if !required {
			return nil
		}
		for _, locale := range domain.SupportedLocales {
			if texts[locale] == "" {
				return fmt.Errorf("%s: cannot be blank", locale)
			}
		}
		return nil
	}
}

func hasText(texts map[string]string) bool {
	for _, text := range texts {
		if text != "" {
			return true
		}
	}
	return false
}
