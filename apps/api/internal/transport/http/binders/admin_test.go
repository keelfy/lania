package binders

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/utils"
)

func TestBindTransferProfileOwner(t *testing.T) {
	profileID := uuid.New()

	tests := []struct {
		name        string
		profileID   string
		body        string
		wantEmail   string
		wantErr     int
		wantInvalid bool
	}{
		{"trims the email", profileID.String(), `{"email":"  New@X.io "}`, "New@X.io", 0, false},
		{"profile id is not a uuid", "abc", `{"email":"a@x.io"}`, "", http.StatusBadRequest, false},
		{"broken json", profileID.String(), `{`, "", http.StatusBadRequest, false},
		{"missing email fails validation", profileID.String(), `{}`, "", 0, true},
		{"invalid email fails validation", profileID.String(), `{"email":"not-an-email"}`, "not-an-email", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.body))
			req.SetPathValue(ProfileIDVariable, tt.profileID)

			cmd, err := BindTransferProfileOwner(req)
			if tt.wantErr != 0 {
				if got := utils.MapCustomErrorToHttpStatus(err); err == nil || got != tt.wantErr {
					t.Fatalf("got error %v with status %d, want status %d", err, got, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if cmd.ProfileID != profileID || cmd.Email != tt.wantEmail {
				t.Errorf("got %+v", cmd)
			}
			if invalid := cmd.Validate() != nil; invalid != tt.wantInvalid {
				t.Errorf("validation failed = %v, want %v", invalid, tt.wantInvalid)
			}
		})
	}
}

func TestBindSaveSeason_ShellAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		address  string
		wantAddr *string
	}{
		{name: "omitted is empty", address: ""},
		{name: "blank is empty", address: `,"shellAddress":"  "`},
		{name: "value is trimmed", address: `,"shellAddress":" shell:9090 "`, wantAddr: stringPointer("shell:9090")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body := `{"name":" Lania V ","startDate":"2026-10-09","preregistration":true,"freeRegistration":true` + tt.address + `}`
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			cmd, err := BindSaveSeason(req)
			if err != nil {
				t.Fatal(err)
			}
			if (cmd.ShellAddress == nil) != (tt.wantAddr == nil) || (tt.wantAddr != nil && *cmd.ShellAddress != *tt.wantAddr) {
				t.Fatalf("shell address = %v, want %v", cmd.ShellAddress, tt.wantAddr)
			}
			if cmd.Name != "Lania V" || !cmd.Preregistration || !cmd.FreeRegistration || cmd.Validate() != nil {
				t.Fatalf("got invalid command: %+v", cmd)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}

func TestBindGrantProduct(t *testing.T) {
	profileID, productID, seasonID := uuid.New(), uuid.New(), uuid.New()
	body := `{"productId":"` + productID.String() + `","seasonId":"` + seasonID.String() + `"}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.SetPathValue(ProfileIDVariable, profileID.String())
	cmd, err := BindGrantProduct(req)
	if err != nil {
		t.Fatal(err)
	}
	if cmd.ProfileID != profileID || cmd.ProductID != productID || cmd.SeasonID != seasonID || cmd.Validate() != nil {
		t.Errorf("got %+v", cmd)
	}

	for name, body := range map[string]string{
		"missing season":  `{"productId":"` + productID.String() + `"}`,
		"missing product": `{"seasonId":"` + seasonID.String() + `"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.SetPathValue(ProfileIDVariable, profileID.String())
		cmd, err := BindGrantProduct(req)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if cmd.Validate() == nil {
			t.Errorf("%s: validation passed", name)
		}
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"productId":"nope"}`))
	req.SetPathValue(ProfileIDVariable, profileID.String())
	if _, err := BindGrantProduct(req); utils.MapCustomErrorToHttpStatus(err) != http.StatusBadRequest {
		t.Errorf("got error %v, want bad request for a product id that is not a uuid", err)
	}
}

func TestBindRevokeGrant(t *testing.T) {
	profileID, grantID := uuid.New(), uuid.New()

	bind := func(grantType string) (*commands.RevokeGrantCommand, error) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		req.SetPathValue(ProfileIDVariable, profileID.String())
		req.SetPathValue(GrantTypeVariable, grantType)
		req.SetPathValue(GrantIDVariable, grantID.String())
		return BindRevokeGrant(req)
	}

	for _, grantType := range []string{"access", "name-color", "name-prefix"} {
		cmd, err := bind(grantType)
		if err != nil || cmd.Validate() != nil || cmd.GrantID != grantID || string(cmd.GrantType) != grantType {
			t.Errorf("%s: got %+v, %v", grantType, cmd, err)
		}
	}

	cmd, err := bind("cape")
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Validate() == nil {
		t.Error("an unknown grant type passed validation")
	}
}

func TestBindGrantCosmetic(t *testing.T) {
	profileID, itemID, seasonID := uuid.New(), uuid.New(), uuid.New()

	tests := []struct {
		name        string
		body        string
		wantSeason  *uuid.UUID
		wantInvalid bool
	}{
		{"name color for good", `{"type":"name-color","itemId":"` + itemID.String() + `"}`, nil, false},
		{"name color for a season", `{"type":"name-color","itemId":"` + itemID.String() + `","seasonId":"` + seasonID.String() + `"}`, &seasonID, false},
		{"special prefix", `{"type":"name-prefix","itemId":"` + itemID.String() + `","prefixType":"special"}`, nil, false},
		{"prefix without a type", `{"type":"name-prefix","itemId":"` + itemID.String() + `"}`, nil, true},
		{"prefix with a wrong type", `{"type":"name-prefix","itemId":"` + itemID.String() + `","prefixType":"shiny"}`, nil, true},
		{"access is no cosmetic", `{"type":"access","itemId":"` + itemID.String() + `"}`, nil, true},
		{"missing type", `{"itemId":"` + itemID.String() + `"}`, nil, true},
		{"missing item", `{"type":"name-color"}`, nil, true},
		{"zero season", `{"type":"name-color","itemId":"` + itemID.String() + `","seasonId":"00000000-0000-0000-0000-000000000000"}`, &uuid.Nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			req.SetPathValue(ProfileIDVariable, profileID.String())

			cmd, err := BindGrantCosmetic(req)
			if err != nil {
				t.Fatal(err)
			}
			if cmd.ProfileID != profileID || (tt.wantSeason == nil) != (cmd.SeasonID == nil) || (tt.wantSeason != nil && *tt.wantSeason != *cmd.SeasonID) {
				t.Errorf("got %+v", cmd)
			}
			if invalid := cmd.Validate() != nil; invalid != tt.wantInvalid {
				t.Errorf("validation failed = %v, want %v", invalid, tt.wantInvalid)
			}
		})
	}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`))
	req.SetPathValue(ProfileIDVariable, profileID.String())
	if _, err := BindGrantCosmetic(req); utils.MapCustomErrorToHttpStatus(err) != http.StatusBadRequest {
		t.Errorf("broken json: got %v, want bad request", err)
	}
}
