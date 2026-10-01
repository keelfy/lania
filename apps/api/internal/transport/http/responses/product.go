package responses

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID       `json:"id"`
	Price       float64         `json:"price"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Metadata    json.RawMessage `json:"metadata"`
	SoldCount   int64           `json:"soldCount"`
	CreatedAt   time.Time       `json:"createdAt"`
}

// ProductCatalog is one page of the shop. NextCursor is empty on the last page.
type ProductCatalog struct {
	Content    []*Product       `json:"content"`
	NextCursor string           `json:"nextCursor,omitempty"`
	Counts     map[string]int64 `json:"counts"`
}
