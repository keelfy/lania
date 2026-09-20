package services

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
	ory "github.com/ory/client-go"
)

// fakeOry serves identities the way Kratos does: by login identifier or in pages with a next token.
type fakeOry struct {
	clients.OryAPI
	identities  []ory.Identity
	byLogin     map[string][]ory.Identity
	pageSize    int
	listCalls   int
	notFoundGet bool
}

func (f *fakeOry) ListIdentities(_ context.Context, pageSize int64, pageToken, credentialsIdentifier string) ([]ory.Identity, string, error) {
	f.listCalls++
	if credentialsIdentifier != "" {
		return f.byLogin[credentialsIdentifier], "", nil
	}

	size := f.pageSize
	if size == 0 {
		size = int(pageSize)
	}
	start, _ := strconv.Atoi(pageToken)
	end := min(start+size, len(f.identities))
	next := ""
	if end < len(f.identities) {
		next = strconv.Itoa(end)
	}
	return f.identities[start:end], next, nil
}

func (f *fakeOry) GetIdentity(context.Context, string) (*ory.Identity, error) {
	if f.notFoundGet {
		return nil, clients.ErrIdentityNotFound
	}
	return nil, errors.New("unexpected call")
}

func identity(email string, role string) ory.Identity {
	identity := ory.Identity{
		Id:     uuid.NewString(),
		Traits: map[string]any{"email": email},
	}
	if role != "" {
		identity.MetadataPublic = map[string]any{"role": role}
	}
	return identity
}

func TestListUsersWithoutSearchKeepsCursor(t *testing.T) {
	fake := &fakeOry{identities: []ory.Identity{identity("a@x.io", ""), identity("b@x.io", "admin"), identity("c@x.io", "")}}
	svc := NewAdminUserService(fake, nil)

	users, next, err := svc.ListUsers(context.Background(), "", "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 || next != "2" {
		t.Fatalf("got %d users and token %q, want 2 users and token 2", len(users), next)
	}
	if users[1].Role != domain.RoleAdmin {
		t.Errorf("got role %q, want admin", users[1].Role)
	}

	users, next, err = svc.ListUsers(context.Background(), "", next, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || next != "" {
		t.Fatalf("got %d users and token %q, want the last page", len(users), next)
	}
}

func TestListUsersSearch(t *testing.T) {
	exact := identity("pw@x.io", "")
	fake := &fakeOry{
		identities: []ory.Identity{identity("google.person@x.io", ""), identity("other@x.io", ""), exact},
		byLogin:    map[string][]ory.Identity{"pw@x.io": {exact}},
	}
	svc := NewAdminUserService(fake, nil)

	t.Run("login identifier hit skips the scan", func(t *testing.T) {
		fake.listCalls = 0
		users, next, err := svc.ListUsers(context.Background(), " PW@x.io ", "ignored", 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 1 || users[0].Email != "pw@x.io" || next != "" {
			t.Fatalf("got %+v and token %q", users, next)
		}
		if fake.listCalls != 1 {
			t.Errorf("got %d list calls, want 1", fake.listCalls)
		}
	})

	t.Run("partial email falls back to a scan", func(t *testing.T) {
		users, next, err := svc.ListUsers(context.Background(), "google", "", 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 1 || users[0].Email != "google.person@x.io" || next != "" {
			t.Fatalf("got %+v and token %q", users, next)
		}
	})

	t.Run("scan stops at the limit", func(t *testing.T) {
		users, _, err := svc.ListUsers(context.Background(), "@x.io", "", 2)
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 2 {
			t.Fatalf("got %d users, want 2", len(users))
		}
	})
}

func TestFindUserByEmail(t *testing.T) {
	password := identity("pw@x.io", "")
	google := identity("Social@X.io", "")
	twin1, twin2 := identity("twin@x.io", ""), identity("twin@x.io", "")
	fake := &fakeOry{
		identities: []ory.Identity{google, twin1, twin2, password},
		byLogin:    map[string][]ory.Identity{"pw@x.io": {password}},
		pageSize:   1,
	}
	svc := NewAdminUserService(fake, nil)

	tests := []struct {
		name       string
		email      string
		wantStatus int
		wantID     string
	}{
		{"password user by login identifier", "PW@x.io", 0, password.Id},
		{"social user by scan, case insensitive", " social@x.io ", 0, google.Id},
		{"unknown email", "nobody@x.io", http.StatusNotFound, ""},
		{"same email on two identities", "twin@x.io", http.StatusConflict, ""},
		{"empty email", "  ", http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := svc.FindUserByEmail(context.Background(), tt.email)
			if tt.wantStatus != 0 {
				if got := utils.MapCustomErrorToHttpStatus(err); err == nil || got != tt.wantStatus {
					t.Fatalf("got error %v with status %d, want status %d", err, got, tt.wantStatus)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if user.ID.String() != tt.wantID {
				t.Errorf("got user %s, want %s", user.ID, tt.wantID)
			}
		})
	}
}

func TestScanStopsAfterMaxPages(t *testing.T) {
	fake := &fakeOry{pageSize: 1}
	for i := 0; i < identityScanMaxPages+10; i++ {
		fake.identities = append(fake.identities, identity("user"+strconv.Itoa(i)+"@x.io", ""))
	}
	svc := NewAdminUserService(fake, nil)

	if _, err := svc.FindUserByEmail(context.Background(), "missing@x.io"); utils.MapCustomErrorToHttpStatus(err) != http.StatusNotFound {
		t.Fatalf("got error %v, want not found", err)
	}
	// one login identifier lookup plus the capped scan
	if want := 1 + identityScanMaxPages; fake.listCalls != want {
		t.Errorf("got %d list calls, want %d", fake.listCalls, want)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	svc := NewAdminUserService(&fakeOry{notFoundGet: true}, nil)

	_, err := svc.GetUserByID(context.Background(), uuid.New())
	if utils.MapCustomErrorToHttpStatus(err) != http.StatusNotFound {
		t.Fatalf("got error %v, want not found", err)
	}
}
