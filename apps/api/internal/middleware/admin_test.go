package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ory "github.com/ory/client-go"
)

func TestAdminOnly(t *testing.T) {
	tests := []struct {
		name     string
		session  *ory.Session
		wantCode int
	}{
		{"no session", nil, http.StatusForbidden},
		{"session without identity", &ory.Session{}, http.StatusForbidden},
		{"identity without role", &ory.Session{Identity: &ory.Identity{}}, http.StatusForbidden},
		{"player role", sessionWithRole("player"), http.StatusForbidden},
		{"moderator role", sessionWithRole("mod"), http.StatusForbidden},
		{"unknown role", sessionWithRole("root"), http.StatusForbidden},
		{"role is not a string", &ory.Session{Identity: &ory.Identity{MetadataPublic: map[string]any{"role": true}}}, http.StatusForbidden},
		{"admin role", sessionWithRole("admin"), http.StatusOK},
		{"owner role", sessionWithRole("owner"), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := AdminOnly()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tt.session != nil {
				req = req.WithContext(withSession(req.Context(), tt.session))
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func sessionWithRole(role string) *ory.Session {
	return &ory.Session{Identity: &ory.Identity{
		Id:             "6f9619ff-8b86-d011-b42d-00c04fc964ff",
		MetadataPublic: map[string]any{"role": role},
	}}
}
