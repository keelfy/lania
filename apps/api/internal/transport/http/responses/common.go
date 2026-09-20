package responses

type Paginated[T any] struct {
	Content       []T   `json:"content"`
	Page          int   `json:"page"`
	Size          int   `json:"size"`
	TotalPages    int64 `json:"totalPages"`
	TotalElements int64 `json:"totalElements"`
}

type CursorPaginated[T any] struct {
	Content       []T   `json:"content"`
	TotalElements int64 `json:"totalElements"`
}

// TokenPaginated is a page of a source that only knows the token of the next page, not the total.
// NextPageToken is empty on the last page.
type TokenPaginated[T any] struct {
	Content       []T    `json:"content"`
	NextPageToken string `json:"nextPageToken,omitempty"`
}

type SearchHit[T any] struct {
	ID     string  `json:"id"`
	Source T       `json:"source"`
	Score  float64 `json:"score"`
}
