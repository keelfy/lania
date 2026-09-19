package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/services"
	storesql "github.com/lania-smp/backend/internal/storage/main"
)

type userError struct {
	Status  int
	Message string
}

func (e *userError) Error() string { return e.Message }

type Server struct {
	Store    *Store
	Identity *IdentityClient
	Shell    clients.ShellAPI
	AdminIDs map[string]bool
	Origin   string
}

type page struct {
	Actor         string    `json:"actor"`
	Search        string    `json:"search"`
	Owner         string    `json:"owner"`
	Next          string    `json:"next"`
	Previous      string    `json:"previous"`
	ActiveSeason  string    `json:"activeSeason"`
	AttachProfile *Profile  `json:"attachProfile"`
	Candidate     *Account  `json:"candidate"`
	Accounts      []Account `json:"accounts"`
	Account       *Account  `json:"account"`
	Profiles      []Profile `json:"profiles"`
	Profile       *Profile  `json:"profile"`
	Products      []Product `json:"products"`
	Seasons       []Season  `json:"seasons"`
	Grants        []Grant   `json:"grants"`
}

type actorKey struct{}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.accounts)
	mux.HandleFunc("GET /profiles", s.profiles)
	mux.HandleFunc("GET /profiles/{id}", s.profile)
	mux.HandleFunc("POST /profiles/{id}/owner", s.owner)
	mux.HandleFunc("POST /profiles/{id}/give", s.give)
	mux.HandleFunc("POST /profiles/{id}/revoke", s.revoke)
	mux.HandleFunc("POST /profiles/{id}/sync", s.sync)
	mux.HandleFunc("GET /session", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"actor": r.Context().Value(actorKey{})})
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Keep same-origin form POSTs from acquiring a null Origin in browsers.
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		if r.URL.Path == "/healthz" && r.Method == http.MethodGet {
			w.WriteHeader(200)
			return
		}
		// Origin is mandatory on every mutation, including same-site sibling subdomains.
		// No state changes use GET, and no CORS permissions are granted.
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if r.Header.Get("Origin") != s.Origin {
				s.fail(w, r, &userError{403, "Источник запроса не разрешён. Откройте админку заново."})
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
			if err := r.ParseForm(); err != nil {
				s.fail(w, r, &userError{400, "Некорректная форма."})
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		actor, err := s.Identity.Session(ctx, r.Header.Get("Cookie"))
		if err != nil {
			s.fail(w, r, err)
			return
		}
		if !s.AdminIDs[actor] {
			s.fail(w, r, &userError{403, "У этого аккаунта нет доступа к админке."})
			return
		}
		mux.ServeHTTP(w, r.WithContext(context.WithValue(ctx, actorKey{}, actor)))
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("admin response", "error", err)
	}
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, data page) {
	data.Actor, _ = r.Context().Value(actorKey{}).(string)
	writeJSON(w, 200, data)
}

func (s *Server) fail(w http.ResponseWriter, r *http.Request, err error) {
	status, message := 500, "Не удалось выполнить операцию. Обновите страницу и повторите."
	var public *userError
	if errors.As(err, &public) {
		status, message = public.Status, public.Message
	} else if errors.Is(err, sql.ErrNoRows) {
		status, message = 404, "Запись не найдена. Обновите страницу."
	} else {
		slog.Error("admin request", "path", r.URL.Path, "error", err)
	}
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) accounts(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	data := page{Search: search}
	attachID := r.URL.Query().Get("attach_to")
	if attachID != "" {
		id, err := parseID(attachID)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		data.AttachProfile, err = s.Store.Profile(r.Context(), id)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		if data.AttachProfile.Owner != nil {
			s.fail(w, r, &userError{409, "Профиль уже привязан к аккаунту. Обновите страницу профиля."})
			return
		}
	}
	var token string
	var err error
	if id, parseErr := parseID(search); parseErr == nil {
		account, lookupErr := s.Identity.Account(r.Context(), id.String())
		if lookupErr != nil {
			var public *userError
			if !errors.As(lookupErr, &public) || public.Status != 404 {
				s.fail(w, r, lookupErr)
				return
			}
		} else {
			data.Accounts = []Account{*account}
		}
	} else {
		data.Accounts, token, err = s.Identity.Accounts(r.Context(), r.URL.Query().Get("page_token"), search)
		if err != nil {
			s.fail(w, r, err)
			return
		}
	}
	if token != "" {
		data.Next = "/?" + url.Values{"q": {search}, "page_token": {token}, "attach_to": {attachID}}.Encode()
	}
	if r.URL.Query().Get("page_token") != "" {
		data.Previous = "/?" + url.Values{"q": {search}, "attach_to": {attachID}}.Encode()
	}
	s.respond(w, r, data)
}

func (s *Server) profiles(w http.ResponseWriter, r *http.Request) {
	search, owner := strings.TrimSpace(r.URL.Query().Get("q")), r.URL.Query().Get("owner")
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if r.URL.Query().Get("offset") == "" {
		offset, err = 0, nil
	}
	if err != nil || offset < 0 || offset > 1000000 {
		s.fail(w, r, &userError{400, "Некорректная страница."})
		return
	}
	data := page{Search: search, Owner: owner}
	if offset > 0 {
		data.Previous = "/profiles?" + url.Values{"q": {search}, "owner": {owner}, "offset": {strconv.Itoa(max(0, offset-50))}}.Encode()
	}
	if owner != "" {
		if _, err := parseID(owner); err != nil {
			s.fail(w, r, err)
			return
		}
		data.Account, err = s.Identity.Account(r.Context(), owner)
		if err != nil {
			s.fail(w, r, err)
			return
		}
	}
	data.Profiles, err = s.Store.Profiles(r.Context(), search, owner, offset)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if len(data.Profiles) > 50 {
		data.Profiles = data.Profiles[:50]
		data.Next = "/profiles?" + url.Values{"q": {search}, "owner": {owner}, "offset": {strconv.Itoa(offset + 50)}}.Encode()
	}
	s.respond(w, r, data)
}

func parseID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, &userError{400, "Укажите корректный UUID."}
	}
	return id, nil
}

func (s *Server) profile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		s.fail(w, r, err)
		return
	}
	data := page{ActiveSeason: s.Store.ActiveSeason.String()}
	data.Profile, err = s.Store.Profile(r.Context(), id)
	if err == nil {
		data.Products, err = s.Store.Products(r.Context())
	}
	if err == nil {
		data.Seasons, err = s.Store.Seasons(r.Context())
	}
	if err == nil {
		data.Grants, err = s.Store.Grants(r.Context(), data.Profile)
	}
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if data.Profile.Owner != nil {
		data.Account, err = s.Identity.Account(r.Context(), data.Profile.OwnerID())
		// Deleted Kratos accounts must not prevent an administrator from detaching a profile.
		if err != nil {
			var public *userError
			if !errors.As(err, &public) || public.Status != 404 {
				s.fail(w, r, err)
				return
			}
		}
	}
	if data.Profile.Owner == nil && r.URL.Query().Get("candidate") != "" {
		candidate, parseErr := parseID(r.URL.Query().Get("candidate"))
		if parseErr != nil {
			s.fail(w, r, parseErr)
			return
		}
		data.Candidate, err = s.Identity.Account(r.Context(), candidate.String())
		if err != nil {
			s.fail(w, r, err)
			return
		}
	}
	s.respond(w, r, data)
}

func (s *Server) mutation(w http.ResponseWriter, r *http.Request, action func(uuid.UUID, uuid.UUID) error, sync bool) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		s.fail(w, r, err)
		return
	}
	actor := uuid.MustParse(r.Context().Value(actorKey{}).(string))
	if err := action(id, actor); err != nil {
		s.fail(w, r, err)
		return
	}
	slog.Info("admin mutation", "actor", actor, "profile", id, "action", r.URL.Path,
		"product", r.PostForm.Get("product"), "grant", r.PostForm.Get("grant"), "season", r.PostForm.Get("season"), "owner", r.PostForm.Get("owner"), "previous_owner", r.PostForm.Get("expected_owner"))
	result := "saved"
	if sync {
		if err := s.syncProfile(r.Context(), id); err != nil {
			slog.Error("admin Minecraft sync", "profile", id, "error", err)
			result = "sync_failed"
		}
	}
	writeJSON(w, 200, map[string]string{"result": result, "profileId": id.String()})
}

func (s *Server) owner(w http.ResponseWriter, r *http.Request) {
	s.mutation(w, r, func(id, actor uuid.UUID) error {
		owner := strings.TrimSpace(r.PostForm.Get("owner"))
		if owner != "" {
			target, err := parseID(owner)
			if err != nil {
				return err
			}
			owner = target.String()
			if _, err := s.Identity.Account(r.Context(), owner); err != nil {
				return err
			}
		}
		return s.Store.Owner(r.Context(), id, actor, r.PostForm.Get("expected_owner"), owner)
	}, false)
}

func (s *Server) give(w http.ResponseWriter, r *http.Request) {
	s.mutation(w, r, func(id, actor uuid.UUID) error {
		product, err := parseID(r.PostForm.Get("product"))
		if err != nil {
			return err
		}
		season, err := parseID(r.PostForm.Get("season"))
		if err != nil {
			return err
		}
		return s.Store.Give(r.Context(), id, product, season, actor)
	}, true)
}

func (s *Server) revoke(w http.ResponseWriter, r *http.Request) {
	s.mutation(w, r, func(id, actor uuid.UUID) error {
		if r.PostForm.Get("confirm") != "yes" {
			return &userError{400, "Подтвердите отзыв продукта."}
		}
		grant, err := parseID(r.PostForm.Get("grant"))
		if err != nil {
			return err
		}
		return s.Store.Revoke(r.Context(), id, grant, actor, r.PostForm.Get("kind"))
	}, true)
}

func (s *Server) sync(w http.ResponseWriter, r *http.Request) {
	s.mutation(w, r, func(uuid.UUID, uuid.UUID) error { return nil }, true)
}

// Re-read desired state while holding the profile lock. A retry always applies current state.
// SQL commits before RPC: an RPC failure is visible and can be retried without granting twice.
func (s *Server) syncProfile(ctx context.Context, id uuid.UUID) error {
	return s.Store.mutate(ctx, id, func(tx *sql.Tx, p *Profile) error {
		q := storesql.WithMySQLTx(tx)
		profile, err := q.FindProfileByID(ctx, id)
		if err != nil {
			return err
		}
		access, err := q.CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, profile.MinecraftUUID, s.Store.ActiveSeason)
		if err != nil {
			return err
		}
		if access {
			err = s.Shell.AddToWhitelist(ctx, profile.MinecraftUUID, profile.MinecraftUsername)
		} else {
			err = s.Shell.RemoveFromWhitelist(ctx, profile.MinecraftUUID)
		}
		if err != nil {
			return err
		}
		prefixes, err := q.FindProfilePrefixesByProfileID(ctx, id)
		if err != nil {
			return err
		}
		var glyph, special *domain.NamePrefix
		for _, prefix := range prefixes {
			switch prefix.Type {
			case domain.ProfilePrefixTypeGlyth:
				glyph = prefix.NamePrefix
			case domain.ProfilePrefixTypeSpecial:
				special = prefix.NamePrefix
			}
		}
		formatted := services.NewProfileCosmeticsService(nil).GetProfileFullPrefix(ctx, profile.NameColor, glyph, special)
		return s.Shell.SetPlayerPrefix(ctx, profile.MinecraftUUID, formatted)
	})
}
