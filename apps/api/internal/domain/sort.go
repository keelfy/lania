package domain

type CursorSort struct {
	Column    string
	Direction string
	Cursor    any
	Limit     int
}

type Sort struct {
	Column    string
	Direction string
}

type Filters = map[string]string

type Pagination struct {
	From int
	Size int
	Page int
}

type Paginated[T any] struct {
	Content       []*T
	Page          int
	Size          int
	TotalPages    int64
	TotalElements int64
}

type CursorPaginated[T any] struct {
	Content       []*T
	Cursor        *CursorSort
	TotalElements int64
}

type SearchHit[T any] struct {
	ID     string
	Source T
	Score  float64
}
