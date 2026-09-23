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
		SessionKey:  "session-key",
		CSRFToken:   "csrf-token",
		Name:        "Лес",
		Description: "Зелёный градиент",
		PriceName:   domain.ProductPriceNameNameColor,
	}
}

func httpStatus(err error) int {
	if customErr, ok := err.(*utils.CustomError); ok {
		return customErr.HttpStatus
	}
	return 0
}

func TestEasyDonateService_CreateShopProduct(t *testing.T) {
	t.Run("happy path returns the new product id", func(t *testing.T) {
		var capturedCookie, capturedCSRFHeader string
		var capturedFields map[string]string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				fmt.Fprint(w, `<div data-product-id="1"></div>`)
				return
			}
			capturedCookie = r.Header.Get("Cookie")
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
		if capturedCookie != "easydonate_session=session-key" {
			t.Errorf("cookie = %q", capturedCookie)
		}
		if capturedCSRFHeader != "csrf-token" {
			t.Errorf("X-CSRF-TOKEN = %q", capturedCSRFHeader)
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
				fmt.Fprint(w, `<div data-product-id="1"></div>`)
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
				fmt.Fprint(w, `<div data-product-id="1"></div>`)
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
				fmt.Fprint(w, `<div data-product-id="1"></div>`)
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
				fmt.Fprint(w, `<div data-product-id="1"></div>`)
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
				fmt.Fprint(w, `<div data-product-id="1"></div>`)
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

	t.Run("tariff without a RUB price is a bad request", func(t *testing.T) {
		service := newEDService(t, "http://unused.invalid", []*domain.ProductPrice{{Name: domain.ProductPriceNameNameColor, Currency: domain.CurrencyUSD, Amount: 2}})
		_, err := service.CreateShopProduct(context.Background(), validEDCommand())
		if httpStatus(err) != http.StatusBadRequest {
			t.Fatalf("err = %v, want bad request", err)
		}
	})
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
