package middleware

import (
	"net/http"

	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
)

// AdminOnly lets only identities with an admin site role pass. It must run after SessionMiddleware.
func AdminOnly() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			session, err := getSession(r.Context())
			if err != nil || session.Identity == nil {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			if !domain.RoleFromMetadata(session.Identity.MetadataPublic).IsAdmin() {
				logger.Warnf(r.Context(), "Identity %s is not an admin", session.Identity.Id)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
