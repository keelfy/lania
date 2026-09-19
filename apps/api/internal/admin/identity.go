package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Only the fields needed by the panel are decoded; credentials never reach the UI.
type Account struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	Traits    struct {
		Email string          `json:"email"`
		Name  json.RawMessage `json:"name"`
	} `json:"traits"`
}

type IdentityClient struct {
	PublicURL, AdminURL string
	HTTP                *http.Client
}

func NewIdentityClient(publicURL, adminURL string) *IdentityClient {
	return &IdentityClient{publicURL, adminURL, &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (c *IdentityClient) get(ctx context.Context, endpoint, cookie string, result any) (http.Header, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.Header, resp.StatusCode, fmt.Errorf("Kratos returned HTTP %d", resp.StatusCode)
	}
	err = json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(result)
	return resp.Header, resp.StatusCode, err
}

func (c *IdentityClient) Session(ctx context.Context, cookie string) (string, error) {
	if cookie == "" {
		return "", &userError{401, "Войдите через аккаунт Lania."}
	}
	var session struct {
		Active   bool    `json:"active"`
		Identity Account `json:"identity"`
	}
	_, status, err := c.get(ctx, c.PublicURL+"/sessions/whoami", cookie, &session)
	if status == 401 || status == 403 {
		return "", &userError{401, "Сессия истекла. Войдите снова."}
	}
	if err != nil {
		return "", err
	}
	if !session.Active || session.Identity.ID == "" || session.Identity.State != "active" {
		return "", &userError{401, "Сессия неактивна. Войдите снова."}
	}
	return session.Identity.ID, nil
}

func (c *IdentityClient) Account(ctx context.Context, id string) (*Account, error) {
	var account Account
	_, status, err := c.get(ctx, c.AdminURL+"/admin/identities/"+url.PathEscape(id), "", &account)
	if status == 404 {
		return nil, &userError{404, "Аккаунт не найден."}
	}
	return &account, err
}

func (c *IdentityClient) Accounts(ctx context.Context, token, email string) ([]Account, string, error) {
	query := url.Values{"page_size": {"50"}}
	if token != "" {
		query.Set("page_token", token)
	}
	if email != "" {
		query.Set("credentials_identifier", email)
	}
	var accounts []Account
	headers, _, err := c.get(ctx, c.AdminURL+"/admin/identities?"+query.Encode(), "", &accounts)
	if err != nil {
		return nil, "", err
	}
	// Follow only the opaque token, never a URL supplied by the upstream.
	var next string
	for _, link := range strings.Split(headers.Get("Link"), ",") {
		parts := strings.Split(link, ";")
		for _, part := range parts[1:] {
			if strings.TrimSpace(part) == `rel="next"` || strings.TrimSpace(part) == "rel=next" {
				u, parseErr := url.Parse(strings.Trim(strings.TrimSpace(parts[0]), "<>"))
				if parseErr == nil {
					next = u.Query().Get("page_token")
				}
			}
		}
	}
	return accounts, next, nil
}
