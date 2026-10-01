package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProductCursorRoundTrip(t *testing.T) {
	product := &Product{
		ID:        uuid.New(),
		Category:  ProductCategoryNamePrefix,
		CreatedAt: time.Date(2026, 9, 1, 12, 30, 0, 0, time.UTC),
	}

	got, err := DecodeProductCursor(NewProductCursor(product).Encode())
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != product.ID || got.Category != product.Category || !got.CreatedAt.Equal(product.CreatedAt) {
		t.Errorf("cursor changed in the round trip: %+v", got)
	}
}

func TestDecodeProductCursorRejectsGarbage(t *testing.T) {
	for name, token := range map[string]string{
		"not base64":      "%%%",
		"not json":        "bm90LWpzb24",
		"missing id":      "eyJjIjoidXBncmFkZSJ9",
		"missing content": "e30",
	} {
		if _, err := DecodeProductCursor(token); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
