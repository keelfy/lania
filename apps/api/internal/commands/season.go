package commands

import (
	"errors"
	"net"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
)

type SaveSeasonCommand struct {
	ID              uuid.UUID
	SeasonNumber    int
	Name            string
	PreviewImage    *string
	StartDate       time.Time
	EndDate         *time.Time
	ServerIP        *string
	ServerPort      *uint16
	IsActive        bool
	RCONPassword    *string
	SetRCONPassword bool
}

func (c *SaveSeasonCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.SeasonNumber, validation.Required, validation.Min(1)),
		validation.Field(&c.Name, validation.Required, validation.RuneLength(1, 255)),
		validation.Field(&c.StartDate, validation.Required),
		validation.Field(&c.EndDate, validation.By(func(value any) error {
			if c.EndDate != nil && c.EndDate.Before(c.StartDate) {
				return errors.New("cannot be before start date")
			}
			return nil
		})),
		validation.Field(&c.ServerIP, validation.By(func(value any) error {
			if c.ServerIP != nil && net.ParseIP(*c.ServerIP) == nil {
				return errors.New("must be a valid IPv4 or IPv6 address")
			}
			if (c.ServerIP == nil) != (c.ServerPort == nil) {
				return errors.New("server IP and port must be set together")
			}
			return nil
		})),
		validation.Field(&c.RCONPassword, validation.RuneLength(0, 4096)),
	)
}
