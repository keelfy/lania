package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const testActor = "11111111-1111-4111-8111-111111111111"

func TestAuthorizationAndCSRF(t *testing.T) {
	var calls atomic.Int64
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path == "/sessions/whoami" {
			switch r.Header.Get("Cookie") {
			case "session=expired":
				http.Error(w, "expired", 401)
			case "session=inactive":
				fmt.Fprintf(w, `{"active":false,"identity":{"id":%q,"state":"active"}}`, testActor)
			case "session=disabled":
				fmt.Fprintf(w, `{"active":true,"identity":{"id":%q,"state":"inactive"}}`, testActor)
			default:
				fmt.Fprintf(w, `{"active":true,"identity":{"id":%q,"state":"active"}}`, testActor)
			}
			return
		}
		if r.URL.Path != "/admin/identities" {
			t.Errorf("unexpected upstream path: %s", r.URL.Path)
		}
		if r.Header.Get("Cookie") != "" {
			t.Error("user cookie leaked to admin API")
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer upstream.Close()
	s := &Server{Identity: NewIdentityClient(upstream.URL, upstream.URL), Origin: "https://admin.example", LoginURL: "https://accounts.example/login", AdminIDs: map[string]bool{testActor: true}}
	for _, tc := range []struct {
		name, method, path, cookie, origin string
		allowed                            bool
		status                             int
	}{
		{"anonymous", "GET", "/", "", "", true, 401},
		{"expired", "GET", "/", "session=expired", "", true, 401},
		{"inactive session", "GET", "/", "session=inactive", "", true, 401},
		{"disabled account", "GET", "/", "session=disabled", "", true, 401},
		{"regular user", "GET", "/", "session=active", "", false, 403},
		{"admin", "GET", "/", "session=active", "", true, 200},
		{"missing origin", "POST", "/profiles/x/give", "session=active", "", true, 403},
		{"sibling origin", "POST", "/profiles/x/give", "session=active", "https://www.example", true, 403},
		{"forged origin", "POST", "/profiles/x/give", "session=active", "https://admin.example.evil", true, 403},
		{"authorized invalid id", "POST", "/profiles/x/give", "session=active", "https://admin.example", true, 400},
		{"regular user mutation", "POST", "/profiles/x/give", "session=active", "https://admin.example", false, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s.AdminIDs[testActor] = tc.allowed
			r := httptest.NewRequest(tc.method, tc.path, nil)
			r.Header.Set("Cookie", tc.cookie)
			r.Header.Set("Origin", tc.origin)
			w := httptest.NewRecorder()
			before := calls.Load()
			s.Handler().ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status %d, expected %d: %s", w.Code, tc.status, w.Body.String())
			}
			if tc.status == 403 && tc.method == "POST" && tc.origin != s.Origin && calls.Load() != before {
				t.Error("CSRF request reached upstream")
			}
			if w.Header().Get("Referrer-Policy") != "same-origin" {
				t.Error("browser form Origin may be suppressed")
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Error("admin response may be cached")
			}
		})
	}
}

func TestAccountsPaginationAndProjection(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page_size") != "50" || r.URL.Query().Get("page_token") != "opaque&token" || r.URL.Query().Get("credentials_identifier") != "a+b@example.com" {
			t.Errorf("wrong query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Link", `<http://internal/admin/identities?page_token=next%2Btoken>; rel="next"`)
		fmt.Fprintf(w, `[{"id":%q,"state":"active","traits":{"email":"a+b@example.com"},"credentials":{"password":{"secret":"never-render-this"}}}]`, testActor)
	}))
	defer upstream.Close()
	accounts, next, err := NewIdentityClient(upstream.URL, upstream.URL).Accounts(context.Background(), "opaque&token", "a+b@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if next != "next+token" || len(accounts) != 1 {
		t.Fatalf("accounts=%v next=%q", accounts, next)
	}
	raw, _ := json.Marshal(accounts)
	if strings.Contains(string(raw), "never-render-this") {
		t.Error("credential retained")
	}
}

func TestAccountUUIDLookup(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sessions/whoami":
			fmt.Fprintf(w, `{"active":true,"identity":{"id":%q,"state":"active"}}`, testActor)
		case "/admin/identities/" + testActor:
			fmt.Fprintf(w, `{"id":%q,"state":"active","traits":{"email":"keeper@example.com"}}`, testActor)
		default:
			t.Errorf("unexpected upstream path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()
	server := &Server{Identity: NewIdentityClient(upstream.URL, upstream.URL), AdminIDs: map[string]bool{testActor: true}}
	r := httptest.NewRequest("GET", "/?q="+testActor, nil)
	r.Header.Set("Cookie", "session=active")
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "keeper@example.com") {
		t.Fatalf("lookup failed: %d %s", w.Code, w.Body.String())
	}
}

func TestJSONErrorResponse(t *testing.T) {
	server := &Server{}
	w := httptest.NewRecorder()
	server.fail(w, httptest.NewRequest("POST", "/profiles/"+testActor+"/owner", nil), &userError{409, "Владелец изменился."})
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != 409 || body["error"] != "Владелец изменился." || w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("wrong error: %d %v", w.Code, body)
	}
}
