package middleware

import (
	"context"
	"net/http"
	"slices"

	"github.com/lania-smp/backend/internal/domain"
)

const CurrencyCtxKey = "req.currency"

func CurrencyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			currency := domain.DefaultCurrency
			currencyQueryParam := r.URL.Query().Get("currency")

			if currencyQueryParam != "" {
				if slices.Contains(domain.AllowedCurrencies, domain.Currency(currencyQueryParam)) {
					currency = domain.Currency(currencyQueryParam)
				}
			}

			ctx := context.WithValue(r.Context(), CurrencyCtxKey, currency)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(hfn)
	}
}
