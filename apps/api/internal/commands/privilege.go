package commands

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// privilegePermissionPattern is the shape of a permission node: lowercase words joined by dots, * allowed.
var privilegePermissionPattern = regexp.MustCompile(`^[a-z0-9_.*-]{1,200}$`)

// reservedPermissionPrefixes start nodes that change groups, weights or chat decorations. A privilege must not
// hand out a role by mistake. The shell refuses them as well.
var reservedPermissionPrefixes = []string{"group.", "weight.", "prefix.", "suffix.", "meta.", "displayname."}

// maxPrivilegePrice is the largest amount the decimal(10, 2) price column holds.
const maxPrivilegePrice = 99999999.99

// ValidatePrivilegePermission checks that the node can be sold as a privilege.
func ValidatePrivilegePermission(permission string) error {
	if !privilegePermissionPattern.MatchString(permission) {
		return errors.New("must be 1 to 200 characters of a-z, 0-9, _, ., * and -")
	}
	if permission == "*" {
		return errors.New("cannot be the wildcard of every permission")
	}
	for _, prefix := range reservedPermissionPrefixes {
		if strings.HasPrefix(permission, prefix) {
			return fmt.Errorf("cannot start with %s", prefix)
		}
	}
	return nil
}

type PrivilegePrice struct {
	Currency domain.Currency
	Amount   float64
}

type SavePrivilegeCommand struct {
	ID   uuid.UUID
	Name string
	// Names translates Name. A missing locale shows Name.
	Names      domain.CosmeticNames
	Permission string
	// Prices holds one price for every allowed currency.
	Prices []PrivilegePrice
}

func (c *SavePrivilegeCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required, validation.Length(1, 255)),
		validation.Field(&c.Names, validCosmeticNames),
		validation.Field(&c.Permission, validation.Required, validation.By(func(value any) error {
			return ValidatePrivilegePermission(value.(string))
		})),
		validation.Field(&c.Prices, validation.By(func(value any) error {
			return validatePrivilegePrices(value.([]PrivilegePrice))
		})),
	)
}

func validatePrivilegePrices(prices []PrivilegePrice) error {
	seen := map[domain.Currency]bool{}
	for _, price := range prices {
		if !slices.Contains(domain.AllowedCurrencies, price.Currency) {
			return fmt.Errorf("currency %q is not supported", price.Currency)
		}
		if seen[price.Currency] {
			return fmt.Errorf("%s is listed twice", price.Currency)
		}
		seen[price.Currency] = true
		if math.IsNaN(price.Amount) || price.Amount <= 0 || price.Amount > maxPrivilegePrice {
			return fmt.Errorf("%s price must be above 0 and at most %.2f", price.Currency, maxPrivilegePrice)
		}
		if math.Abs(price.Amount*100-math.Round(price.Amount*100)) > 1e-6 {
			return fmt.Errorf("%s price must have at most two decimals", price.Currency)
		}
	}
	for _, currency := range domain.AllowedCurrencies {
		if !seen[currency] {
			return fmt.Errorf("%s price is required", currency)
		}
	}
	return nil
}
