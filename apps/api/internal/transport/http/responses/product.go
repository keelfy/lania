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
