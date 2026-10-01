package domain

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProductCursor is the position of the last product of a catalog page. The catalog is ordered by
// category, then creation time, then id, all descending, so the three together are a unique key.
type ProductCursor struct {
	Category  ProductCategory `json:"c"`
	CreatedAt time.Time       `json:"t"`
	ID        uuid.UUID       `json:"i"`
}

func NewProductCursor(product *Product) *ProductCursor {
	return &ProductCursor{Category: product.Category, CreatedAt: product.CreatedAt, ID: product.ID}
}

// Encode returns the opaque token that clients send back to get the next page.
func (c *ProductCursor) Encode() string {
	payload, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(payload)
}

func DecodeProductCursor(token string) (*ProductCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var cursor ProductCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return nil, err
	}
	if cursor.Category == "" || cursor.ID == uuid.Nil {
		return nil, errors.New("incomplete product cursor")
	}
	return &cursor, nil
}

type ProductPage struct {
	Products   []*Product
	NextCursor *ProductCursor
}
