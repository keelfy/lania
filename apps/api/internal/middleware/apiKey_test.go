package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiKey(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	tests := []struct {
		name       string
		configured string
		sent       string
		want       int
	}{
		{"right key", "secret", "secret", http.StatusOK},
		{"wrong key", "secret", "other", http.StatusForbidden},
		{"no key sent", "secret", "", http.StatusForbidden},
		{"no key configured", "", "", http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("API_KEY", tt.configured)
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.sent != "" {
				req.Header.Set("x-api-key", tt.sent)
			}
			rec := httptest.NewRecorder()
			ApiKey()(next).ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
