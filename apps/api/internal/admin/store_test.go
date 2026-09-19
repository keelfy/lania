package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
)

// Each integration test creates and drops its own database on the provided test server.
func testStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("ADMIN_TEST_DSN")
	if dsn == "" {
		t.Skip("set ADMIN_TEST_DSN to run MariaDB integration tests")
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.DBName = ""
	cfg.ParseTime = true
	cfg.MultiStatements = true
	root, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	name := "admin_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := root.Exec("CREATE DATABASE " + name); err != nil {
		root.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = root.Exec("DROP DATABASE " + name); root.Close() })
	cfg.DBName = name
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	schema, err := os.ReadFile("../../db/migration/20250208103928_init.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	return &Store{DB: db, ActiveSeason: uuid.New(), DefaultColor: uuid.MustParse("2628bf9d-5b7c-438b-900a-67753261a823")}
}

func TestStoreProductAndOwnerLifecycle(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	actor, owner, profile, mc, color, prefix := uuid.MustParse(testActor), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := s.DB.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	count := func(query string, args ...any) int {
		t.Helper()
		var n int
		if err := s.DB.QueryRow(query, args...).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	exec(`INSERT INTO seasons (id,season_number,start_date) VALUES (?,1,now())`, s.ActiveSeason)
	exec(`INSERT INTO profiles (id,mc_uuid,mc_username,role,is_slim) VALUES (?,?,'Tester','player',false)`, profile, mc)
	exec(`INSERT INTO name_colors (id,name,colors) VALUES (?,'Red','{"colors":["#ff0000"]}')`, color)
	exec(`INSERT INTO name_prefixes (id,name,metadata) VALUES (?,'Star','{"prefix":"*"}')`, prefix)
	products := []struct {
		id                 uuid.UUID
		category, metadata string
	}{
		{uuid.New(), "upgrade", `{"action":"season_access"}`},
		{uuid.New(), "name-color", fmt.Sprintf(`{"nameColorId":%q}`, color)},
		{uuid.New(), "name-prefix", fmt.Sprintf(`{"namePrefixId":%q}`, prefix)},
	}
	for _, p := range products {
		exec(`INSERT INTO products (id,category,metadata,price_name) VALUES (?,?,?,'test')`, p.id, p.category, p.metadata)
	}
	if err := s.Owner(ctx, profile, actor, "", owner.String()); err != nil {
		t.Fatal(err)
	}
	if err := s.Owner(ctx, profile, actor, "", ""); err == nil {
		t.Fatal("stale owner update accepted")
	}
	if err := s.Owner(ctx, profile, actor, owner.String(), uuid.NewString()); err == nil {
		t.Fatal("silent transfer accepted")
	}
	if err := s.Owner(ctx, profile, actor, owner.String(), ""); err != nil {
		t.Fatal(err)
	}
	loaded, err := s.Profile(ctx, profile)
	if err != nil || loaded.Owner != nil {
		t.Fatalf("detach: %v, %v", loaded, err)
	}
	for _, p := range products {
		for range 2 {
			if err := s.Give(ctx, profile, p.id, s.ActiveSeason, actor); err != nil {
				t.Fatal(err)
			}
		}
	}
	grants, err := s.Grants(ctx, loaded)
	if err != nil {
		t.Fatal(err)
	}
	if len(grants) != 3 {
		t.Fatalf("duplicate grants: %v", grants)
	}
	if err := s.Give(ctx, profile, products[0].id, uuid.New(), actor); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing season: %v", err)
	}
	profiles, err := s.Profiles(ctx, "Tester", "", 0)
	if err != nil || len(profiles) != 1 {
		t.Fatalf("search: %v %v", profiles, err)
	}
	if _, err := s.Products(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Seasons(ctx); err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE profiles SET name_color_id = ? WHERE id = ?`, color, profile)
	exec(`INSERT INTO profile_prefixes (profile_id,type,name_prefix_id) VALUES (?,'glyth',?)`, profile, prefix)
	permanent := uuid.New()
	exec(`INSERT INTO profile_name_color_options (id,profile_id,name_color_id) VALUES (?,?,?)`, permanent, profile, color)
	var colorGrant, prefixGrant, accessGrant uuid.UUID
	for _, g := range grants {
		switch g.Kind {
		case "color":
			colorGrant = uuid.MustParse(g.ID)
		case "prefix":
			prefixGrant = uuid.MustParse(g.ID)
		case "access":
			accessGrant = uuid.MustParse(g.ID)
		}
	}
	if err := s.Revoke(ctx, profile, colorGrant, actor, "color"); err != nil {
		t.Fatal(err)
	}
	if count(`SELECT COUNT(*) FROM profiles WHERE id = ? AND name_color_id = ?`, profile, color) != 1 {
		t.Fatal("valid permanent color reset")
	}
	if err := s.Revoke(ctx, profile, permanent, actor, "color"); err != nil {
		t.Fatal(err)
	}
	if count(`SELECT COUNT(*) FROM profiles WHERE id = ? AND name_color_id = ?`, profile, s.DefaultColor) != 1 {
		t.Fatal("revoked selected color retained")
	}
	if err := s.Revoke(ctx, profile, prefixGrant, actor, "prefix"); err != nil {
		t.Fatal(err)
	}
	if count(`SELECT COUNT(*) FROM profile_prefixes WHERE profile_id = ?`, profile) != 0 {
		t.Fatal("revoked selected prefix retained")
	}
	exec(`INSERT INTO profile_accesses (mc_uuid,season_id,source) VALUES (?,?,'free')`, mc, s.ActiveSeason)
	if err := s.Revoke(ctx, profile, accessGrant, actor, "access"); err != nil {
		t.Fatal(err)
	}
	if count(`SELECT COUNT(*) FROM profile_accesses WHERE mc_uuid = ?`, mc) != 0 {
		t.Fatal("duplicate access retained")
	}
	defaultGrant := uuid.New()
	exec(`INSERT INTO profile_name_color_options (id,profile_id,name_color_id) VALUES (?,?,?)`, defaultGrant, profile, s.DefaultColor)
	if err := s.Revoke(ctx, profile, defaultGrant, actor, "color"); err == nil {
		t.Fatal("default color revoked")
	}
	otherProfile := uuid.New()
	exec(`INSERT INTO profiles (id,mc_uuid,mc_username,role,is_slim) VALUES (?,?,'Other','player',false)`, otherProfile, uuid.New())
	if err := s.Revoke(ctx, otherProfile, defaultGrant, actor, "color"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-profile revoke: %v", err)
	}
}

type shellRecorder struct {
	clients.ShellAPI
	fail    bool
	allowed bool
	prefix  string
}

func (s *shellRecorder) AddToWhitelist(context.Context, uuid.UUID, string) error {
	if s.fail {
		return errors.New("offline")
	}
	s.allowed = true
	return nil
}
func (s *shellRecorder) RemoveFromWhitelist(context.Context, uuid.UUID) error {
	if s.fail {
		return errors.New("offline")
	}
	s.allowed = false
	return nil
}
func (s *shellRecorder) SetPlayerPrefix(_ context.Context, _ uuid.UUID, prefix string) error {
	if s.fail {
		return errors.New("offline")
	}
	s.prefix = prefix
	return nil
}

func TestHTTPMutationPersistsAndRetriesMinecraftSync(t *testing.T) {
	store := testStore(t)
	profile, mc, product := uuid.New(), uuid.New(), uuid.New()
	for _, query := range []string{
		fmt.Sprintf(`INSERT INTO seasons (id,season_number,start_date) VALUES ('%s',1,now())`, store.ActiveSeason),
		fmt.Sprintf(`INSERT INTO profiles (id,mc_uuid,mc_username,role,is_slim) VALUES ('%s','%s','Tester','player',false)`, profile, mc),
		fmt.Sprintf(`INSERT INTO products (id,category,metadata,price_name) VALUES ('%s','upgrade','{"action":"season_access"}','access')`, product),
	} {
		if _, err := store.DB.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"active":true,"identity":{"id":%q,"state":"active"}}`, testActor)
	}))
	defer upstream.Close()
	shell := &shellRecorder{fail: true}
	s := &Server{Store: store, Shell: shell, Identity: NewIdentityClient(upstream.URL, upstream.URL), AdminIDs: map[string]bool{testActor: true}, Origin: "https://admin.example"}
	post := func(action string, form url.Values) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("POST", "/profiles/"+profile.String()+"/"+action, strings.NewReader(form.Encode()))
		r.Header.Set("Cookie", "session=active")
		r.Header.Set("Origin", s.Origin)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	response := post("give", url.Values{"product": {product.String()}, "season": {store.ActiveSeason.String()}})
	if response.Code != 200 || !strings.Contains(response.Body.String(), "sync_failed") {
		t.Fatalf("failure response: %d %s", response.Code, response.Body.String())
	}
	var grants int
	if err := store.DB.QueryRow(`SELECT COUNT(*) FROM profile_accesses WHERE mc_uuid = ?`, mc).Scan(&grants); err != nil || grants != 1 {
		t.Fatalf("desired state lost: %d %v", grants, err)
	}
	shell.fail = false
	response = post("sync", url.Values{})
	if response.Code != 200 || !strings.Contains(response.Body.String(), "saved") || !shell.allowed || shell.prefix != "<reset>" {
		t.Fatalf("retry failed: %d %s %+v", response.Code, response.Body.String(), shell)
	}
	var grant string
	if err := store.DB.QueryRow(`SELECT id FROM profile_accesses WHERE mc_uuid = ?`, mc).Scan(&grant); err != nil {
		t.Fatal(err)
	}
	response = post("revoke", url.Values{"grant": {grant}, "kind": {"access"}})
	if response.Code != 400 {
		t.Fatal("missing confirmation accepted")
	}
	response = post("revoke", url.Values{"grant": {grant}, "kind": {"access"}, "confirm": {"yes"}})
	if response.Code != 200 || shell.allowed {
		t.Fatalf("revocation failed: %d %+v", response.Code, shell)
	}
}

func TestAccountSelectionPreservesContext(t *testing.T) {
	store := testStore(t)
	profile, mc := uuid.New(), uuid.New()
	if _, err := store.DB.Exec(`INSERT INTO profiles (id,mc_uuid,mc_username,role,is_slim) VALUES (?,?,'Picker','player',false)`, profile, mc); err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sessions/whoami":
			fmt.Fprintf(w, `{"active":true,"identity":{"id":%q,"state":"active"}}`, testActor)
		case "/admin/identities":
			w.Header().Set("Link", `<http://internal/admin/identities?page_token=next>; rel="next"`)
			fmt.Fprintf(w, `[{"id":%q,"state":"active","traits":{"email":"keeper@example.com"}}]`, testActor)
		case "/admin/identities/" + testActor:
			fmt.Fprintf(w, `{"id":%q,"state":"active","traits":{"email":"keeper@example.com"}}`, testActor)
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()
	server := &Server{Store: store, Identity: NewIdentityClient(upstream.URL, upstream.URL), AdminIDs: map[string]bool{testActor: true}, Origin: "https://admin.example"}
	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("Cookie", "session=active")
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		return w
	}
	w := get("/?attach_to=" + profile.String())
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"username":"Picker"`) || !strings.Contains(w.Body.String(), testActor) || !strings.Contains(w.Body.String(), "attach_to="+profile.String()+`\u0026page_token=next`) {
		t.Fatalf("selection context lost: %d %s", w.Code, w.Body.String())
	}
	w = get("/profiles/" + profile.String() + "?candidate=" + testActor)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"candidate":{`) || !strings.Contains(w.Body.String(), `"id":"`+testActor+`"`) {
		t.Fatalf("candidate preview missing: %d %s", w.Code, w.Body.String())
	}
	// Selection links must become conflicts once the owner changes.
	if err := store.Owner(context.Background(), profile, uuid.MustParse(testActor), "", testActor); err != nil {
		t.Fatal(err)
	}
	if w = get("/?attach_to=" + profile.String()); w.Code != 409 {
		t.Fatalf("stale selection accepted: %d", w.Code)
	}
}
