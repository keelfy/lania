package binders

import (
	"net/http"
	"strconv"

	"github.com/lania-smp/backend/internal/domain"
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
