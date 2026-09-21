package commands

import (
	"errors"
	"net"
	"net/url"
	"strconv"
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
	ShellAddress     *string
	PlanURL          *string
	IsActive         bool
	IsPrimary        bool
	Preregistration  bool
	FreeRegistration bool
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
		validation.Field(&c.ShellAddress, validation.By(validateShellAddress(c.ShellAddress))),
		validation.Field(&c.PlanURL, validation.By(validatePlanURL(c.PlanURL))),
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

// validateShellAddress requires host:port, the form gRPC dials.
func validateShellAddress(address *string) validation.RuleFunc {
	return func(value any) error {
		if address == nil {
			return nil
		}
		host, portValue, err := net.SplitHostPort(*address)
		if err != nil {
			return errors.New("must be host:port")
		}
		port, err := strconv.ParseUint(portValue, 10, 16)
		if err != nil || port == 0 {
			return errors.New("port must be between 1 and 65535")
		}
		if !isServerAddress(host) {
			return errors.New("must be a valid IP address or hostname")
		}
		return nil
	}
}

// validatePlanURL requires an absolute http(s) URL, the form a browser link opens.
func validatePlanURL(planURL *string) validation.RuleFunc {
	return func(value any) error {
		if planURL == nil {
			return nil
		}
		if len(*planURL) > 2048 {
			return errors.New("must be at most 2048 characters")
		}
		parsed, err := url.Parse(*planURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
			return errors.New("must be an http or https URL")
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
