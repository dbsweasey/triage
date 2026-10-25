package store

import (
	_ "embed"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestGetArticle(t *testing.T) {
	s := createTestDb(t)

	want := Article{
		Title: "Set up multi-factor authentication (MFA)",
		Body:  "Install the company-approved authenticator app on your phone. Sign in to the security portal, choose \"Add sign-in method\", and scan the QR code. Approve the test notification to finish. If you get a new phone, register it before wiping the old one, or ask the help desk to reset your MFA methods.",
		Tags:  []string{"account", "mfa", "2fa", "security", "authenticator"},
	}

	got, err := s.GetArticle(t.Context(), 3)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if got.Title != want.Title {
		t.Errorf("title = %v, want %v", got.Title, want.Title)
	}
	if got.Body != want.Body {
		t.Errorf("body = %v, want %v", got.Body, want.Body)
	}
	if !slices.Equal(got.Tags, want.Tags) {
		t.Errorf("tags = %v, want %v", got.Tags, want.Tags)
	}

	got, err = s.GetArticle(t.Context(), 200)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want %v", err, ErrNotFound)
	}
}

func TestListArticles(t *testing.T) {
	s := createTestDb(t)

	var articles []Article
	if err := json.Unmarshal(seedKB, &articles); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := "Access a shared drive or folder"

	got, err := s.ListArticles(t.Context())
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if got == nil {
		t.Fatalf("list = nil, want [...]")
	}
	if len(got) != len(articles) {
		t.Fatalf("len(list) = %d, want %d", len(got), len(articles))
	}
	if !slices.IsSortedFunc(got, func(a, b Article) int {
		return strings.Compare(a.Title, b.Title)
	}) {
		t.Errorf("list not sorted by title")
	}
	if got[0].Title != want {
		t.Errorf("Title = %q, want %q", got[0].Title, want)
	}
}

func TestSearchArticles(t *testing.T) {
	tests := []struct {
		Name    string
		Query   string
		Limit   int
		Want    string
		WantLen int
	}{
		{Name: "normal", Query: "My VPN is disconnecting every 15 minutes", Want: "VPN keeps disconnecting", Limit: 5, WantLen: 5},
		{Name: "punctuation", Query: "I CAn't print!!", Want: "Printer shows offline or won't print", Limit: 5, WantLen: 1},
		{Name: "zero limit", Query: "My VPN won't stop disconnecting", Want: "VPN keeps disconnecting", Limit: 0, WantLen: 5},
		{Name: "one limit", Query: "My VPN keeps disconnecting", Want: "VPN keeps disconnecting", Limit: 1, WantLen: 1},
		{Name: "no results", Query: "xxyzyz?", Want: "", WantLen: 0},
		{Name: "stopwords", Query: "my is it he am are", Want: "", WantLen: 0},
		{Name: "nasty", Query: `vpn" OR NEAR( wifi *`, Want: "VPN keeps disconnecting", WantLen: 3},
	}

	s := createTestDb(t)

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := s.SearchArticles(t.Context(), tc.Query, tc.Limit)
			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if got == nil {
				t.Fatalf("list = nil, want non-nil")
			}
			if len(got) != tc.WantLen {
				t.Errorf("len(list) = %d, want %d", len(got), tc.WantLen)
			}
			if tc.WantLen > 0 && got[0].Title != tc.Want {
				t.Errorf("title = %q, want %q", got[0].Title, tc.Want)
			}
		})
	}
}

func TestBuildFTSQuery(t *testing.T) {
	tests := []struct {
		Name string
		Text string
		Want string
	}{
		{Name: "lowercase", Text: "my vpn keeps disconnecting", Want: `"vpn" OR "keeps" OR "disconnecting"`},
		{Name: "mixed case", Text: "My VPN keeps disconnecting", Want: `"vpn" OR "keeps" OR "disconnecting"`},
		{Name: "punctuation", Text: "My VPN won't stop disconnecting!!!", Want: `"vpn" OR "wont" OR "stop" OR "disconnecting"`},
		{Name: "all stop words", Text: "is a i AM a The", Want: ""},
		{Name: "empty", Text: "", Want: ""},
		{Name: "meaningful short words", Text: "pc ip io os bug", Want: `"pc" OR "ip" OR "io" OR "os" OR "bug"`},
		{Name: "operators", Text: "printer NEAR scanner", Want: `"printer" OR "near" OR "scanner"`},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			got := buildFTSQuery(tc.Text)
			if tc.Want != got {
				t.Errorf("query = %q, want %q", got, tc.Want)
			}
		})
	}
}

func TestSeeding(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	s1, err := OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	t.Cleanup(func() { s1.Close() })
	list, err := s1.ListArticles(t.Context())
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	length := len(list)
	if length == 0 {
		t.Fatalf("len(list) = 0, want > 0")
	}

	s1.Close()

	s2, err := OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	list, err = s2.ListArticles(t.Context())
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if length != len(list) {
		t.Errorf("len(list) = %d, want %d", len(list), length)
	}

	t.Cleanup(func() { s2.Close() })
}
