package clients

import (
	"net/http"
	"testing"
)

func TestNextPageToken(t *testing.T) {
	tests := []struct {
		name string
		link []string
		want string
	}{
		{"no header", nil, ""},
		{"first and next", []string{`</admin/identities?page_size=2&page_token=first>; rel="first", </admin/identities?page_size=2&page_token=abc>; rel="next"`}, "abc"},
		{"separate header values", []string{`<https://kratos/admin/identities?page_size=2&page_token=first>; rel="first"`, `<https://kratos/admin/identities?page_size=2&page_token=xyz>; rel="next"`}, "xyz"},
		{"last page has no next", []string{`</admin/identities?page_size=2&page_token=first>; rel="first"`}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{}
			for _, value := range tt.link {
				header.Add("Link", value)
			}
			if got := nextPageToken(header); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
