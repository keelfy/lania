package binders

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

const (
	UnreadQueryParam = "unread"
	OffsetQueryParam = "offset"
)

// BindNotificationFilter reads ?unread=true&offset=&limit=. A bad number falls back to its default.
func BindNotificationFilter(r *http.Request) domain.NotificationFilter {
	limit, err := strconv.Atoi(BindOptionalQueryParamAsString(r, LimitQueryParam, strconv.Itoa(DefaultLimit)))
	if err != nil {
		limit = DefaultLimit
	}
	offset, err := strconv.Atoi(BindOptionalQueryParamAsString(r, OffsetQueryParam, "0"))
	if err != nil {
		offset = 0
	}
	unreadOnly, _ := strconv.ParseBool(BindOptionalQueryParamAsString(r, UnreadQueryParam, "false"))

	return domain.NotificationFilter{UnreadOnly: unreadOnly, Offset: offset, Limit: limit}
}

func BindSendAnnouncement(r *http.Request) (*commands.SendAnnouncementCommand, error) {
	req := &requests.SendAnnouncement{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}

	return &commands.SendAnnouncementCommand{
		Title: trimLocalizedText(req.Title),
		Body:  trimLocalizedText(req.Body),
		Link:  strings.TrimSpace(req.Link),
	}, nil
}

func trimLocalizedText(texts map[string]string) map[string]string {
	trimmed := make(map[string]string, len(texts))
	for locale, text := range texts {
		trimmed[locale] = strings.TrimSpace(text)
	}
	return trimmed
}
