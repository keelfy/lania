package domain

import (
	"time"

	"github.com/google/uuid"
)

type BasketItem struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ProductID uuid.UUID
	ProfileID uuid.UUID
	Quantity  int
	CreatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedAt time.Time
	UpdatedBy uuid.UUID
	// relations
	Product *Product
	Profile *Profile
}
