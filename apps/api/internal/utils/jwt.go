package utils

import (
	"context"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	ory "github.com/ory/client-go"
)

var ErrJWTMissing = NewForbiddenError("Authorization header is missing", nil)

func GetUserIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	session, ok := ctx.Value("req.session").(*ory.Session)
	if !ok || session == nil {
		return uuid.Nil, ErrJWTMissing
	}
	userID, err := uuid.Parse(session.Identity.Id)
	if err != nil {
		return uuid.Nil, ErrJWTMissing
	}
	return userID, nil
}

func GetUserIDFromContextOrNil(ctx context.Context) *uuid.UUID {
	session, ok := ctx.Value("req.session").(*ory.Session)
	if !ok || session == nil {
		return nil
	}
	userID, err := uuid.Parse(session.Identity.Id)
	if err != nil {
		return nil
	}
	return &userID
}

func GetLocaleFromCtx(ctx context.Context) string {
	locale, ok := ctx.Value("req.locale").(string)
	if !ok || locale == "" {
		return DefaultLocale
	}
	return locale
}

func GetCurrencyFromCtx(ctx context.Context) domain.Currency {
	currency, ok := ctx.Value("req.currency").(domain.Currency)
	if !ok || currency == "" {
		return domain.DefaultCurrency
	}
	return currency
}
