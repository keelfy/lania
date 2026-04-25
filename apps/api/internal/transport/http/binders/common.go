package binders

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

var (
	UserIDVariable        = "userId"
	OrderIDVariable       = "orderId"
	UsernameVariable      = "username"
	ContentNoteIDVariable = "contentNoteId"
	ContentIDVariable     = "contentId"
	CategoryVariable      = "category"
	CollectionIDVariable  = "collectionId"
	ItemIDVariable        = "itemId"
)

var (
	AvatarSizeParam = "avatarSize"
	CoverSizeParam  = "coverSize"
)

func BindPathVariable(r *http.Request, name string) (string, error) {
	value := r.PathValue(name)

	if len(value) == 0 {
		return "", utils.NewBadRequestError(fmt.Sprintf("path variable %v is required", name), nil)
	}

	return value, nil
}

func BindPathVariableAsUUID(r *http.Request, name string) (uuid.UUID, error) {
	value := r.PathValue(name)

	if len(value) == 0 {
		return uuid.Nil, utils.NewBadRequestError(fmt.Sprintf("path variable %v is required", name), nil)
	}

	uid, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, utils.NewBadRequestError(fmt.Sprintf("path variable %v is not a valid UUID", name), err)
	}
	return uid, nil
}

func BindMandatoryQueryParamAsString(r *http.Request, key string) (string, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return "", utils.NewBadRequestError(fmt.Sprintf("query parameter %v is required", key), nil)
	}

	decodedValue, err := url.QueryUnescape(value)
	if err != nil {
		return "", utils.NewBadRequestError(fmt.Sprintf("query parameter %v is invalid", key), err)
	}

	return decodedValue, nil
}

func BindMandatoryQueryParamAsInt64(r *http.Request, key string) (int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0, utils.NewBadRequestError(fmt.Sprintf("query parameter %v is required", key), nil)
	}

	decodedValue, err := url.QueryUnescape(value)
	if err != nil {
		return 0, utils.NewBadRequestError(fmt.Sprintf("query parameter %v is not a valid int64", key), err)
	}

	return strconv.ParseInt(decodedValue, 10, 64)
}

func BindMandatoryQueryParamAsUUID(r *http.Request, key string) (uuid.UUID, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return uuid.Nil, utils.NewBadRequestError(fmt.Sprintf("query parameter %v is required", key), nil)
	}

	decodedValue, err := url.QueryUnescape(value)
	if err != nil {
		return uuid.Nil, utils.NewBadRequestError(fmt.Sprintf("query parameter %v is invalid", key), err)
	}

	return uuid.Parse(decodedValue)
}

func BindOptionalQueryParamAsString(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	decodedValue, err := url.QueryUnescape(value)
	if err != nil {
		return defaultValue
	}
	return decodedValue
}

func BindOptionalQueryParamAsUUID(r *http.Request, key string) (*uuid.UUID, error) {
	value := BindOptionalQueryParamAsString(r, key, "")

	if value == "" {
		return nil, nil
	}

	uid, err := uuid.Parse(value)
	if err != nil {
		return nil, utils.NewBadRequestError(fmt.Sprintf("query parameter %v is not a valid UUID", key), err)
	}
	return &uid, nil
}

func BindOptionalQueryParamAsUUIDs(r *http.Request, key string) (uuid.UUIDs, error) {
	value := BindOptionalQueryParamAsString(r, key, "")
	uuids := uuid.UUIDs{}

	if value == "" {
		return uuids, nil
	}

	for _, id := range strings.Split(value, ",") {
		uid, err := uuid.Parse(id)
		if err != nil {
			return uuids, utils.NewBadRequestError(fmt.Sprintf("query parameter %v is not a valid UUID", key), err)
		}
		uuids = append(uuids, uid)
	}
	return uuids, nil
}

var (
	DefaultLimit = 20
	MinLimit     = 0
	MaxLimit     = 100
	DefaultPage  = 0
)

var (
	PageQueryParam      = "page"
	SizeQueryParam      = "size"
	LimitQueryParam     = "limit"
	ColumnQueryParam    = "column"
	DirectionQueryParam = "direction"
	CursorQueryParam    = "cursor"
	FiltersQueryParam   = "filters"
)

func BindPagination(r *http.Request) (*domain.Pagination, error) {
	page, err := strconv.Atoi(BindOptionalQueryParamAsString(r, PageQueryParam, strconv.Itoa(DefaultPage)))
	if err != nil {
		page = DefaultPage
	}

	size, err := strconv.Atoi(BindOptionalQueryParamAsString(r, SizeQueryParam, strconv.Itoa(DefaultLimit)))
	if err != nil || size <= 0 {
		size = DefaultLimit
	}

	if size > MaxLimit {
		size = MaxLimit
	} else if size < MinLimit {
		size = DefaultLimit
	}

	pagination := &domain.Pagination{
		Size: size,
		Page: page,
		From: page * size,
	}

	return pagination, nil
}

func BindSort(r *http.Request) *domain.Sort {
	column := BindOptionalQueryParamAsString(r, ColumnQueryParam, "")
	direction := BindOptionalQueryParamAsString(r, DirectionQueryParam, "")
	return &domain.Sort{
		Column:    column,
		Direction: direction,
	}
}

// ParseCursor parses the cursor value from query parameters based on the expected column type.
func BindCursor(r *http.Request, columnType string) (any, error) {
	cursorParam := BindOptionalQueryParamAsString(r, CursorQueryParam, "")
	if cursorParam == "" {
		return nil, nil // No cursor provided; handle this as the "first page" case
	}

	decodedCursorParam, err := url.QueryUnescape(cursorParam)
	if err != nil {
		return nil, utils.NewBadRequestError("invalid cursor: expected a valid URL-encoded string", err)
	}

	switch columnType {
	case "int":
		cursor, err := strconv.Atoi(decodedCursorParam)
		if err != nil {
			return nil, utils.NewBadRequestError("invalid cursor: expected an integer", err)
		}
		return cursor, nil
	case "float":
		cursor, err := strconv.ParseFloat(decodedCursorParam, 64)
		if err != nil {
			return nil, utils.NewBadRequestError("invalid cursor: expected a float", err)
		}
		return cursor, nil
	case "datetime":
		cursor, err := time.Parse(time.RFC3339, decodedCursorParam)
		if err != nil {
			return nil, utils.NewBadRequestError("invalid cursor: expected a datetime in RFC3339 format", err)
		}
		return cursor, nil
	default:
		return decodedCursorParam, nil
	}
}

func BindPaginatedCursorSort(r *http.Request) (*domain.CursorSort, error) {
	limitParam := BindOptionalQueryParamAsString(r, LimitQueryParam, strconv.Itoa(DefaultLimit))

	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		limit = DefaultLimit
	}

	if limit > MaxLimit {
		limit = MaxLimit
	} else if limit < MinLimit {
		limit = DefaultLimit
	}

	cursor, err := BindCursor(r, "string")
	if err != nil {
		return nil, err
	}

	column := BindOptionalQueryParamAsString(r, ColumnQueryParam, "created_at")
	direction := BindOptionalQueryParamAsString(r, DirectionQueryParam, "desc")

	sort := &domain.CursorSort{
		Cursor:    cursor,
		Limit:     limit,
		Column:    column,
		Direction: direction,
	}
	return sort, nil
}

func BindFilters(r *http.Request) (domain.Filters, error) {
	filters := make(domain.Filters)
	param := BindOptionalQueryParamAsString(r, FiltersQueryParam, "")
	if param == "" {
		return filters, nil
	}

	decodedParam, err := url.QueryUnescape(param)
	if err != nil {
		return nil, utils.NewBadRequestError("invalid filters: expected a valid URL-encoded string", err)
	}

	for filter := range strings.SplitSeq(decodedParam, ",") {
		parts := strings.Split(filter, ":")
		if len(parts) == 2 {
			filters[parts[0]] = parts[1]
		}
	}

	return filters, nil
}
