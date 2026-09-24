package services

import (
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type fakeEDPriceQueries struct {
	sql.Queries
	prices []*domain.ProductPrice
}

func (q *fakeEDPriceQueries) FindPricesByNames(context.Context, []domain.ProductPriceName) ([]*domain.ProductPrice, error) {
	return q.prices, nil
}

func newEDService(t *testing.T, cpURL string, prices []*domain.ProductPrice) *easyDonateService {
	t.Helper()
	t.Setenv("ED_CP_URL", cpURL)
	t.Setenv("ED_SHOP_ID", "42")
	queries := &fakeEDPriceQueries{prices: prices}
	return &easyDonateService{storage: &stubMainStorage{queries: queries}}
}

func rubPrices() []*domain.ProductPrice {
	return []*domain.ProductPrice{{Name: domain.ProductPriceNameNameColor, Currency: domain.CurrencyRUB, Amount: 199}}
}

func validEDCommand() *commands.CreateEDProductCommand {
	return &commands.CreateEDProductCommand{
		UserAuth:    "user-auth",
		Name:        "Лес",
		Description: "Зелёный градиент",
		PriceName:   domain.ProductPriceNameNameColor,
	}
}

// edProductsPageHTML is a products page with one existing row and the CSRF meta tag October renders.
const edProductsPageHTML = `<meta name="csrf-token" content="page-token"><div data-product-id="1"></div>`

func httpStatus(err error) int {
	if customErr, ok := err.(*utils.CustomError); ok {
		return customErr.HttpStatus
	}
	return 0
}

func TestEasyDonateService_CreateShopProduct(t *testing.T) {
	t.Run("happy path returns the new product id", func(t *testing.T) {
		var getCookie, capturedCSRFHeader string
		var postCookies []*http.Cookie
		var capturedFields map[string]string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				getCookie = r.Header.Get("Cookie")
				http.SetCookie(w, &http.Cookie{Name: "easydonate_session", Value: "fresh-session", Path: "/"})
				fmt.Fprint(w, edProductsPageHTML)
				return
			}
			postCookies = r.Cookies()
			capturedCSRFHeader = r.Header.Get("X-CSRF-TOKEN")
			capturedFields = parseMultipartFields(t, r)
			fmt.Fprint(w, `{"products_row": "<div data-product-id=\"1\"></div><div data-product-id=\"2\"></div>"}`)
		}))
		defer server.Close()

		service := newEDService(t, server.URL, rubPrices())
		id, err := service.CreateShopProduct(context.Background(), validEDCommand())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != 2 {
			t.Errorf("id = %d, want 2", id)
		}
		if getCookie != "user_auth=user-auth" {
			t.Errorf("GET cookie = %q", getCookie)
		}
		postCookieValues := map[string]string{}
		for _, cookie := range postCookies {
			postCookieValues[cookie.Name] = cookie.Value
		}
		if postCookieValues["user_auth"] != "user-auth" || postCookieValues["easydonate_session"] != "fresh-session" {
			t.Errorf("POST cookies = %+v", postCookieValues)
		}
		if capturedCSRFHeader != "page-token" || capturedFields["_token"] != "page-token" {
			t.Errorf("X-CSRF-TOKEN = %q, _token = %q", capturedCSRFHeader, capturedFields["_token"])
		}
		if capturedFields["type"] != "group" || capturedFields["commands[0]"] != "lpv user {user} add permission" {
			t.Errorf("fields = %+v", capturedFields)
		}
		if capturedFields["price"] != "199" {
			t.Errorf("price = %q, want 199", capturedFields["price"])
		}
	})

	t.Run("image is attached as a file part", func(t *testing.T) {
		var hasImagePart bool
		var imageFilename string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, edProductsPageHTML)
				return
			}
			_, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			reader := multipart.NewReader(r.Body, params["boundary"])
			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				if part.FormName() == "image" {
					hasImagePart = true
					imageFilename = part.FileName()
				}
			}
			fmt.Fprint(w, `{"products_row": "<div data-product-id=\"1\"></div><div data-product-id=\"2\"></div>"}`)
		}))
		defer server.Close()

		service := newEDService(t, server.URL, rubPrices())
		cmd := validEDCommand()
		cmd.Image = []byte{1, 2, 3}
		cmd.ImageFilename = "preview.png"
		if _, err := service.CreateShopProduct(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasImagePart {
			t.Error("expected an image part")
		}
		if imageFilename != "preview.png" {
			t.Errorf("image filename = %q", imageFilename)
		}
	})

	t.Run("several new ids returns the largest", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, edProductsPageHTML)
				return
			}
			fmt.Fprint(w, `{"products_row": "<div data-product-id=\"3\"></div><div data-product-id=\"9\"></div>"}`)
		}))
		defer server.Close()

		service := newEDService(t, server.URL, rubPrices())
		id, err := service.CreateShopProduct(context.Background(), validEDCommand())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != 9 {
			t.Errorf("id = %d, want 9", id)
		}
	})

	t.Run("no new id is a conflict", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, edProductsPageHTML)
				return
			}
			fmt.Fprint(w, `{"products_row": "<div data-product-id=\"1\"></div>"}`)
		}))
		defer server.Close()

		service := newEDService(t, server.URL, rubPrices())
		_, err := service.CreateShopProduct(context.Background(), validEDCommand())
		if httpStatus(err) != http.StatusConflict {
			t.Fatalf("err = %v, want conflict", err)
		}
	})

	t.Run("login page on the initial GET is unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, `<html><body>login form, no markers here</body></html>`)
				return
			}
			t.Error("POST should not be reached")
		}))
		defer server.Close()

		service := newEDService(t, server.URL, rubPrices())
		_, err := service.CreateShopProduct(context.Background(), validEDCommand())
		if httpStatus(err) != http.StatusUnauthorized {
			t.Fatalf("err = %v, want unauthorized", err)
		}
	})

	t.Run("419 on the POST is unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, edProductsPageHTML)
				return
			}
			w.WriteHeader(419)
		}))
		defer server.Close()

		service := newEDService(t, server.URL, rubPrices())
		_, err := service.CreateShopProduct(context.Background(), validEDCommand())
		if httpStatus(err) != http.StatusUnauthorized {
			t.Fatalf("err = %v, want unauthorized", err)
		}
	})

	t.Run("non-JSON 200 is an internal error carrying the body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, edProductsPageHTML)
				return
			}
			fmt.Fprint(w, `not json at all`)
		}))
		defer server.Close()

		service := newEDService(t, server.URL, rubPrices())
		_, err := service.CreateShopProduct(context.Background(), validEDCommand())
		if httpStatus(err) != http.StatusInternalServerError {
			t.Fatalf("err = %v, want internal error", err)
		}
	})

	t.Run("page without a meta tag falls back to the XSRF-TOKEN cookie", func(t *testing.T) {
		var xsrfHeader, csrfHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				http.SetCookie(w, &http.Cookie{Name: "XSRF-TOKEN", Value: "xsrf%3D%3D", Path: "/"})
				fmt.Fprint(w, `<div data-product-id="1"></div>`)
				return
			}
			xsrfHeader = r.Header.Get("X-XSRF-TOKEN")
			csrfHeader = r.Header.Get("X-CSRF-TOKEN")
			fmt.Fprint(w, `{"products_row": "<div data-product-id=\"2\"></div>"}`)
		}))
		defer server.Close()

		service := newEDService(t, server.URL, rubPrices())
		if _, err := service.CreateShopProduct(context.Background(), validEDCommand()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if xsrfHeader != "xsrf==" || csrfHeader != "" {
			t.Errorf("X-XSRF-TOKEN = %q, X-CSRF-TOKEN = %q", xsrfHeader, csrfHeader)
		}
	})

	t.Run("page without any CSRF token still posts", func(t *testing.T) {
		var fields map[string]string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, `<div data-product-id="1"></div>`)
				return
			}
			fields = parseMultipartFields(t, r)
			fmt.Fprint(w, `{"products_row": "<div data-product-id=\"2\"></div>"}`)
		}))
		defer server.Close()

		service := newEDService(t, server.URL, rubPrices())
		id, err := service.CreateShopProduct(context.Background(), validEDCommand())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != 2 {
			t.Errorf("id = %d, want 2", id)
		}
		if _, ok := fields["_token"]; ok {
			t.Errorf("_token sent without a token on the page")
		}
	})

	t.Run("tariff without a RUB price is a bad request", func(t *testing.T) {
		service := newEDService(t, "http://unused.invalid", []*domain.ProductPrice{{Name: domain.ProductPriceNameNameColor, Currency: domain.CurrencyUSD, Amount: 2}})
		_, err := service.CreateShopProduct(context.Background(), validEDCommand())
		if httpStatus(err) != http.StatusBadRequest {
			t.Fatalf("err = %v, want bad request", err)
		}
	})
}

func TestFindEDCSRFToken(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"meta", `<meta name="csrf-token" content="a1">`, "a1"},
		{"meta reversed with single quotes", `<meta content='a2' name='csrf-token' />`, "a2"},
		{"hidden input", `<form><input type="hidden" name="_token" value="a3"></form>`, "a3"},
		{"html entity", `<meta name="csrf-token" content="a&amp;4">`, "a&4"},
		{"none", `<meta name="viewport" content="width=device-width">`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findEDCSRFToken([]byte(tt.body)); got != tt.want {
				t.Errorf("findEDCSRFToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func parseMultipartFields(t *testing.T, r *http.Request) map[string]string {
	t.Helper()
	if err := r.ParseMultipartForm(1024 * 1024); err != nil {
		t.Fatalf("failed to parse multipart form: %v", err)
	}
	fields := make(map[string]string)
	for key, values := range r.MultipartForm.Value {
		if len(values) > 0 {
			fields[key] = values[0]
		}
	}
	return fields
}
