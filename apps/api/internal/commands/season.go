package commands

import (
	"errors"
	"net"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
)

type SaveSeasonCommand struct {
	ID               uuid.UUID
	Name             string
	PreviewImage     *string
	StartDate        time.Time
	EndDate          *time.Time
	PublicAddress    *string
	SystemAddress    *string
	RCONPort         *uint16
	IsActive         bool
	IsPrimary        bool
	Preregistration  bool
	FreeRegistration bool
	RCONPassword     *string
	SetRCONPassword  bool
}

func (c *SaveSeasonCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required, validation.RuneLength(1, 255)),
		validation.Field(&c.StartDate, validation.Required),
		validation.Field(&c.EndDate, validation.By(func(value any) error {
			if c.EndDate != nil && c.EndDate.Before(c.StartDate) {
				return errors.New("cannot be before start date")
			}
			return nil
		})),
		validation.Field(
			&c.PublicAddress,
			validation.By(validateAddress(c.PublicAddress)),
		),
		validation.Field(
			&c.SystemAddress,
			validation.By(validateAddressAndPort(c.SystemAddress, c.RCONPort, "system address and RCON port")),
		),
		validation.Field(&c.RCONPassword, validation.RuneLength(0, 4096)),
	)
}

func validateAddress(address *string) validation.RuleFunc {
	return func(value any) error {
		if address != nil && !isServerAddress(*address) {
			return errors.New("must be a valid IP address or hostname")
		}
		return nil
	}
}

func validateAddressAndPort(address *string, port *uint16, fieldNames string) validation.RuleFunc {
	return func(value any) error {
		if (address == nil) != (port == nil) {
			return errors.New(fieldNames + " must be set together")
		}
		if port != nil && *port == 0 {
			return errors.New("port must be between 1 and 65535")
		}
		if address != nil && !isServerAddress(*address) {
			return errors.New("must be a valid IP address or hostname")
		}
		return nil
	}
}

func isServerAddress(address string) bool {
	if net.ParseIP(address) != nil {
		return true
	}
	if len(address) == 0 || len(address) > 253 {
		return false
	}
	for _, label := range strings.Split(strings.TrimSuffix(address, "."), ".") {
		isInvalidLength := len(label) == 0 || len(label) > 63
		hasInvalidHyphen := !isInvalidLength && (label[0] == '-' || label[len(label)-1] == '-')
		if isInvalidLength || hasInvalidHyphen {
			return false
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') &&
				(char < '0' || char > '9') && char != '-' {
				return false
			}
		}
	}
	return true
}
