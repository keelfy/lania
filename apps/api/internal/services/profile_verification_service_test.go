package services

import (
	"context"
	stdsql "database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

// fakeVerificationQueries keeps one profile and its request in memory; an expired request is simply absent.
type fakeVerificationQueries struct {
	sql.Queries
	profile      *domain.Profile
	mojangUUID   *uuid.UUID
	verification *domain.ProfileVerification
}

func (q *fakeVerificationQueries) FindProfileByID(_ context.Context, id uuid.UUID) (*domain.Profile, error) {
	if id != q.profile.ID {
		return nil, stdsql.ErrNoRows
	}
	return q.profile, nil
}

func (q *fakeVerificationQueries) FindProfileByUsername(_ context.Context, username string) (*domain.Profile, error) {
	if username != q.profile.MinecraftUsername {
		return nil, stdsql.ErrNoRows
	}
	return q.profile, nil
}

func (q *fakeVerificationQueries) FindProfileMojangUUIDsByMinecraftUUIDs(context.Context, uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error) {
	res := map[uuid.UUID]uuid.UUID{}
	if q.mojangUUID != nil {
		res[q.profile.MinecraftUUID] = *q.mojangUUID
	}
	return res, nil
}

func (q *fakeVerificationQueries) UpsertProfileVerification(_ context.Context, profileID uuid.UUID, ttl time.Duration) error {
	q.verification = &domain.ProfileVerification{ProfileID: profileID, ExpiresAt: time.Now().Add(ttl)}
	return nil
}

func (q *fakeVerificationQueries) FindOpenProfileVerification(context.Context, uuid.UUID) (*domain.ProfileVerification, error) {
	if q.verification == nil {
		return nil, stdsql.ErrNoRows
	}
	copied := *q.verification
	return &copied, nil
}

func (q *fakeVerificationQueries) SetProfileVerificationCode(_ context.Context, _ uuid.UUID, code string, _ time.Duration) error {
	q.verification.Code = &code
	return nil
}

func (q *fakeVerificationQueries) IncrementProfileVerificationAttempts(context.Context, uuid.UUID) error {
	q.verification.Attempts++
	return nil
}

func (q *fakeVerificationQueries) DeleteProfileVerification(context.Context, uuid.UUID) error {
	q.verification = nil
	return nil
}

func (q *fakeVerificationQueries) SetProfileVerified(_ context.Context, _ uuid.UUID, mcUUID uuid.UUID) error {
	if mcUUID == q.profile.MinecraftUUID {
		now := time.Now()
		q.profile.VerifiedAt = &now
	}
	return nil
}

func newVerificationFixture() (*fakeVerificationQueries, ProfileVerificationService, uuid.UUID) {
	owner := uuid.New()
	premium := uuid.New()
	queries := &fakeVerificationQueries{
		profile:    &domain.Profile{ID: uuid.New(), MinecraftUUID: premium, MinecraftUsername: "keelfy", OwnerUserID: &owner},
		mojangUUID: &premium,
	}
	return queries, NewProfileVerificationService(&stubMainStorage{queries: queries}), owner
}

func TestProfileVerification(t *testing.T) {
	ctx := context.Background()

	t.Run("owner verifies with the code shown on join", func(t *testing.T) {
		queries, service, owner := newVerificationFixture()
		profile := queries.profile
		if _, err := service.StartVerification(ctx, owner, profile.ID); err != nil {
			t.Fatalf("StartVerification() error = %v", err)
		}
		code, err := service.IssueLoginCode(ctx, profile.MinecraftUUID, "keelfy", true)
		if err != nil || len(code) != verificationCodeLength {
			t.Fatalf("IssueLoginCode() = %q, %v, want a code", code, err)
		}
		// A second join shows the same code, so the owner can type the one from the first kick.
		if again, _ := service.IssueLoginCode(ctx, profile.MinecraftUUID, "keelfy", true); again != code {
			t.Errorf("second join got %q, want %q", again, code)
		}
		if err := service.ConfirmVerification(ctx, owner, profile.ID, " "+code+" "); err != nil {
			t.Fatalf("ConfirmVerification() error = %v", err)
		}
		if profile.VerifiedAt == nil || queries.verification != nil {
			t.Errorf("verified at %v, request %v; want verified and the request closed", profile.VerifiedAt, queries.verification)
		}
		if _, err := service.StartVerification(ctx, owner, profile.ID); httpStatus(err) != http.StatusConflict {
			t.Errorf("start on a verified profile: status %d, want conflict", httpStatus(err))
		}
	})

	t.Run("no code for an offline login or another UUID", func(t *testing.T) {
		queries, service, owner := newVerificationFixture()
		profile := queries.profile
		if _, err := service.StartVerification(ctx, owner, profile.ID); err != nil {
			t.Fatalf("StartVerification() error = %v", err)
		}
		if code, _ := service.IssueLoginCode(ctx, profile.MinecraftUUID, "keelfy", false); code != "" {
			t.Errorf("offline login got code %q", code)
		}
		if code, _ := service.IssueLoginCode(ctx, uuid.New(), "keelfy", true); code != "" {
			t.Errorf("login with another UUID got code %q", code)
		}
	})

	t.Run("no code without an open request", func(t *testing.T) {
		queries, service, _ := newVerificationFixture()
		if code, err := service.IssueLoginCode(ctx, queries.profile.MinecraftUUID, "keelfy", true); code != "" || err != nil {
			t.Errorf("IssueLoginCode() = %q, %v, want nothing", code, err)
		}
		if code, err := service.IssueLoginCode(ctx, uuid.New(), "stranger", true); code != "" || err != nil {
			t.Errorf("unknown player: IssueLoginCode() = %q, %v, want nothing", code, err)
		}
	})

	t.Run("start needs the owner and a licensed profile", func(t *testing.T) {
		queries, service, owner := newVerificationFixture()
		if _, err := service.StartVerification(ctx, uuid.New(), queries.profile.ID); httpStatus(err) != http.StatusForbidden {
			t.Errorf("stranger: status %d, want forbidden", httpStatus(err))
		}
		queries.mojangUUID = nil
		if _, err := service.StartVerification(ctx, owner, queries.profile.ID); httpStatus(err) != http.StatusBadRequest {
			t.Errorf("free nickname: status %d, want bad request", httpStatus(err))
		}
		other := uuid.New()
		queries.mojangUUID = &other
		if _, err := service.StartVerification(ctx, owner, queries.profile.ID); httpStatus(err) != http.StatusBadRequest {
			t.Errorf("not rekeyed: status %d, want bad request", httpStatus(err))
		}
	})

	t.Run("wrong codes close the request", func(t *testing.T) {
		queries, service, owner := newVerificationFixture()
		profile := queries.profile
		if _, err := service.StartVerification(ctx, owner, profile.ID); err != nil {
			t.Fatalf("StartVerification() error = %v", err)
		}
		// Before the join there is no code, a guess is not counted.
		if err := service.ConfirmVerification(ctx, owner, profile.ID, "AAAAAA"); httpStatus(err) != http.StatusBadRequest || queries.verification.Attempts != 0 {
			t.Errorf("guess before join: status %d, attempts %d", httpStatus(err), queries.verification.Attempts)
		}
		if _, err := service.IssueLoginCode(ctx, profile.MinecraftUUID, "keelfy", true); err != nil {
			t.Fatalf("IssueLoginCode() error = %v", err)
		}
		for i := 1; i < verificationMaxAttempts; i++ {
			if err := service.ConfirmVerification(ctx, owner, profile.ID, "wrong"); httpStatus(err) != http.StatusBadRequest {
				t.Fatalf("attempt %d: status %d, want bad request", i, httpStatus(err))
			}
		}
		if err := service.ConfirmVerification(ctx, owner, profile.ID, "wrong"); httpStatus(err) != http.StatusConflict {
			t.Errorf("last attempt: status %d, want conflict", httpStatus(err))
		}
		if queries.verification != nil || profile.VerifiedAt != nil {
			t.Error("the request stayed open or the profile got verified")
		}
		// Expired and closed requests look the same: nothing open.
		if err := service.ConfirmVerification(ctx, owner, profile.ID, "wrong"); httpStatus(err) != http.StatusConflict {
			t.Errorf("after close: status %d, want conflict", httpStatus(err))
		}
	})
}
