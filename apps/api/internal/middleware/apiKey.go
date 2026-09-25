package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/lania-smp/backend/internal/config"
)

// ApiKey lets through only requests with the x-api-key of API_KEY. Without API_KEY every request is rejected.
func ApiKey() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			expected := config.GetApiKey()
			apiKey := r.Header.Get("x-api-key")
			if expected == "" || subtle.ConstantTimeCompare([]byte(apiKey), []byte(expected)) != 1 {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
