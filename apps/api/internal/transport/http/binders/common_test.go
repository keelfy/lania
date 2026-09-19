package binders

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestBindSearch(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"missing", "", ""},
		{"trimmed", "search=%20%20Steve%20", "Steve"},
		{"percent is not unescaped twice", "search=" + url.QueryEscape("50%"), "50%"},
		{"cut to max length", "search=" + strings.Repeat("a", 40), strings.Repeat("a", MaxSearchLength)},
		{"cut by characters", "search=" + strings.Repeat("я", 40), strings.Repeat("я", MaxSearchLength)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/v1/profiles?"+tt.query, nil)
			if got := BindSearch(r); got != tt.want {
				t.Errorf("BindSearch = %q, want %q", got, tt.want)
			}
		})
	}
}
