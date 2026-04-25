package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/logger"
	ory "github.com/ory/client-go"
)

func SessionMiddleware(oryAPI clients.OryAPI, optional bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			cookies := r.Header.Get("Cookie")

			session, err := oryAPI.GetSession(r.Context(), cookies)
			if (err != nil || session == nil || !*session.Active) && !optional {
				logger.Debugf(r.Context(), "Session %v is not active: %v", session, err)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := withSession(r.Context(), session)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(hfn)
	}
}

func withSession(ctx context.Context, v *ory.Session) context.Context {
	return context.WithValue(ctx, "req.session", v)
}

func getSession(ctx context.Context) (*ory.Session, error) {
	session, ok := ctx.Value("req.session").(*ory.Session)
	if !ok || session == nil {
		return nil, errors.New("session not found in context")
	}
	return session, nil
}
