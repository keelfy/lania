package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ProductCategory string

const (
	ProductCategoryUpgrade    ProductCategory = "upgrade"
	ProductCategoryNameColor  ProductCategory = "name-color"
	ProductCategoryNamePrefix ProductCategory = "name-prefix"
)

type Currency string

const (
	CurrencyRUB Currency = "RUB"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyBRL Currency = "BRL"
	CurrencyTRY Currency = "TRY"
	CurrencyPLN Currency = "PLN"
)

const DefaultCurrency = CurrencyRUB

var AllowedCurrencies = []Currency{
	CurrencyRUB,
	CurrencyUSD,
	CurrencyEUR,
	CurrencyBRL,
	CurrencyTRY,
	CurrencyPLN,
}

type ProductPriceName string

const (
	ProductPriceNameSeasonAccess ProductPriceName = "season_access"
	ProductPriceNameNameColor    ProductPriceName = "name_color"
	ProductPriceNameNamePrefix   ProductPriceName = "name_prefix"
)

type ProductPrice struct {
	Name     ProductPriceName
	Currency Currency
	Amount   float64
}

type Product struct {
	ID        uuid.UUID
	PriceName ProductPriceName
	Category  ProductCategory
	Metadata  json.RawMessage
	SoldCount int64
	CreatedAt time.Time
	UpdatedAt time.Time
	UpdatedBy *uuid.UUID
	// relations
	Prices        []*ProductPrice
	Localizations []*ProductLocalization
}

type ProductLocalization struct {
	ProductID   uuid.UUID
	Locale      string
	Name        string
	Description string
}

type ProductUpgradeAction string

const (
	ProductUpgradeActionSeasonAccess ProductUpgradeAction = "season_access"
)

type UpgradeProductMetadata struct {
	Action ProductUpgradeAction `json:"action"`
}

type NameColorProductMetadata struct {
	NameColorID uuid.UUID `json:"nameColorId"`
	Colors      []string  `json:"colors"`
}

type NamePrefixProductMetadata struct {
	Prefix       string    `json:"prefix"`
	NamePrefixID uuid.UUID `json:"namePrefixId"`
}
