package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"triage/internal/store"
	"triage/internal/triage"
)

type fakeKB struct {
	articles  []store.Article
	lastQuery string
	lastLimit int
}

func initFakeKB() *fakeKB {
	s := fakeKB{articles: []store.Article{
		{ID: 1, Title: "Test article", Body: "This is a test article for testing", Tags: []string{"test", "fake", "kb"}},
	}}
	return &s
}

func (s *fakeKB) GetArticle(ctx context.Context, id int64) (store.Article, error) {
	for _, a := range s.articles {
		if a.ID == id {
			return a, nil
		}
	}
	return store.Article{}, store.ErrNotFound
}

func (s *fakeKB) ListArticles(context.Context) ([]store.Article, error) {
	return s.articles, nil
}

func (s *fakeKB) SearchArticles(ctx context.Context, text string, limit int) ([]store.Article, error) {
	s.lastQuery = text
	s.lastLimit = limit
	return s.articles, nil
}

type fakeTriages struct {
	triages []store.Triage
}

func (s *fakeTriages) GetLatestTriage(ctx context.Context, ticketID int64) (store.Triage, error) {
	if s.triages[0].TicketID == ticketID {
		return s.triages[0], nil
	}
	return store.Triage{}, store.ErrNotFound
}

func (s *fakeTriages) CreateTriage(ctx context.Context, t store.Triage) (store.Triage, error) {
	t.ID = int64(len(s.triages)) + 1
	s.triages = append(s.triages, t)
	return t, nil
}

func (s *fakeTriages) LatestTriages(context.Context) (map[int64]store.Triage, error) {
	latest := make(map[int64]store.Triage)
	for _, tr := range s.triages {
		latest[tr.TicketID] = tr
	}
	return latest, nil
}

type fakeTriager struct{}

func (s *fakeTriager) GetResult(ctx context.Context, t store.Ticket, as []store.Article) (triage.Result, error) {
	if t.Title == "refuse" {
		return triage.Result{}, triage.ErrRefused
	}
	return triage.Result{Priority: "medium", Category: "network", Summary: "internet went out", Steps: []string{"Restart", "Reconnect"}, ArticleIDs: []int64{}}, nil
}

func newTestServer(ts ticketStore, ks kbStore, trs triageStore, tri triager) http.Handler {
	srv := &server{tickets: ts, articles: ks, triages: trs, triager: tri}
	return srv.routes()
}

func TestCreateTicket(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid ticket", `{"title": "x", "description": "y"}`, http.StatusCreated},
		{"missing title", `{"title": "", "description": "y"}`, http.StatusBadRequest},
		{"invalid json", `{"title":`, http.StatusBadRequest},
	}

	mux := newTestServer(store.NewMemoryStore(), initFakeKB(), &fakeTriages{}, &fakeTriager{})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/tickets", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestGetTicket(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantTitle  string
		wantError  bool
	}{
		{"valid id", "/api/tickets/1", http.StatusOK, "test", false},
		{"not found id", "/api/tickets/99", http.StatusNotFound, "", true},
		{"invalid id", "/api/tickets/abc", http.StatusBadRequest, "", true},
	}

	mux := newTestServer(store.NewMemoryStore(), initFakeKB(), &fakeTriages{}, &fakeTriager{})
	createTicket(t, mux, "test") // ID 1

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			if tc.wantError {
				var e struct {
					Error string `json:"error"`
				}
				if err := json.NewDecoder(rec.Body).Decode(&e); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if e.Error == "" {
					t.Errorf("error message is empty, want non-empty")
				}
				return
			}

			if tc.wantTitle == "" {
				return
			}
			var got store.Ticket
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got.Title != tc.wantTitle {
				t.Errorf("title = %q, want %q", got.Title, tc.wantTitle)
			}
		})
	}
}

func TestListTickets(t *testing.T) {
	mux := newTestServer(store.NewMemoryStore(), initFakeKB(), &fakeTriages{}, &fakeTriager{})
	for range 2 {
		createTicket(t, mux, "t")
	}
	createTriage(t, mux, 1)

	req := httptest.NewRequest(http.MethodGet, "/api/tickets", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	var got []ticketListItem
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != 2 {
		t.Errorf("ID = %d, want 2", got[0].ID)
	}
	if got[0].ID == 1 && got[0].Triage == nil {
		t.Errorf("triage = nil, want non-nil")
	}
	if got[1].ID == 2 && got[1].Triage != nil {
		t.Errorf("triage = not nil, want nil")
	}
}

func TestGetArticle(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantTitle  string
		wantID     int64
	}{
		{"valid id", "/api/kb/1", http.StatusOK, "Test article", 1},
		{"not found id", "/api/kb/99", http.StatusNotFound, "", 0},
		{"invalid id", "/api/kb/abc", http.StatusBadRequest, "", 0},
	}

	mux := newTestServer(nil, initFakeKB(), &fakeTriages{}, &fakeTriager{})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			if tc.wantStatus >= 400 {
				var e struct {
					Error string `json:"error"`
				}
				if err := json.NewDecoder(rec.Body).Decode(&e); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if e.Error == "" {
					t.Errorf("error message is empty, want non-empty")
				}
				return
			}

			if tc.wantTitle == "" {
				return
			}
			var got store.Article
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got.Title != tc.wantTitle {
				t.Errorf("title = %q, want %q", got.Title, tc.wantTitle)
			}
			if got.ID != tc.wantID {
				t.Errorf("id = %d, want %d", got.ID, tc.wantID)
			}
		})
	}
}

func TestListArticles(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantLimit  int
		wantStatus int
		wantQuery  string
	}{
		{"search no limit", "/api/kb?q=vpn", 0, http.StatusOK, "vpn"},
		{"search limit", "/api/kb?q=vpn&limit=1", 1, http.StatusOK, "vpn"},
		{"invalid limit", "/api/kb?q=vpn&limit=abc", 0, http.StatusBadRequest, "vpn"},
		{"list", "/api/kb", 0, http.StatusOK, ""},
	}

	kb := initFakeKB()
	mux := newTestServer(store.NewMemoryStore(), kb, &fakeTriages{}, &fakeTriager{})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset fake KB
			kb.lastLimit = 0
			kb.lastQuery = ""
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("code = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantStatus >= 400 {
				return
			}
			var got []store.Article
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(got) == 0 {
				t.Fatalf("len = 0, want > 1")
			}
			if kb.lastQuery != tc.wantQuery {
				t.Errorf("query = %q, want %q", kb.lastQuery, tc.wantQuery)
			}
			if kb.lastLimit != tc.wantLimit {
				t.Errorf("limit = %d, want %d", kb.lastLimit, tc.wantLimit)
			}
			if got[0].ID != 1 {
				t.Errorf("ID = %d, want 1", got[0].ID)
			}
		})
	}
}

func TestCreateTriage(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"no ticket", "/api/tickets/3/triage", http.StatusNotFound},
		{"invalid id", "/api/tickets/abc/triage", http.StatusBadRequest},
		{"valid ticket", "/api/tickets/1/triage", http.StatusCreated},
		{"refused", "/api/tickets/2/triage", http.StatusUnprocessableEntity},
	}
	mux := newTestServer(store.NewMemoryStore(), initFakeKB(), &fakeTriages{}, &fakeTriager{})
	createTicket(t, mux, "test")   // ID 1
	createTicket(t, mux, "refuse") // ID 2

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(""))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestGetLatestTriage(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantID       int64
		wantTicketID int64
		wantStatus   int
	}{
		{"valid", "/api/tickets/1/triage", 1, 1, http.StatusOK},
		{"invalid ticket", "/api/tickets/abc/triage", 0, 0, http.StatusBadRequest},
		{"no ticket", "/api/tickets/3/triage", 0, 0, http.StatusNotFound},
		{"no triage", "/api/tickets/2/triage", 0, 2, http.StatusNotFound},
	}
	mux := newTestServer(store.NewMemoryStore(), initFakeKB(), &fakeTriages{}, &fakeTriager{})
	createTicket(t, mux, "test")  // ID 1
	createTicket(t, mux, "test2") // ID 2
	createTriage(t, mux, 1)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, strings.NewReader(""))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestHealth(t *testing.T) {
	mux := newTestServer(store.NewMemoryStore(), initFakeKB(), &fakeTriages{}, &fakeTriager{})
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
}

func createTicket(t *testing.T, h http.Handler, title string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/tickets", strings.NewReader(`{"title":"`+title+`","description":"d"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup: create status = %d, want 201", rec.Code)
	}
}

func createTriage(t *testing.T, h http.Handler, ticketID int64) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/"+strconv.FormatInt(ticketID, 10)+"/triage", strings.NewReader(""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup: create status = %d, want 201", rec.Code)
	}
}

// Empty struct implementation of ticketStore to test 500 errors
type failingStore struct{}

func (failingStore) Create(context.Context, store.Ticket) (store.Ticket, error) {
	return store.Ticket{}, errors.New("disk on fire!")
}

func (failingStore) Get(context.Context, int64) (store.Ticket, error) {
	return store.Ticket{}, errors.New("disk on fire!")
}

func (failingStore) List(context.Context) ([]store.Ticket, error) {
	return nil, errors.New("disk on fire!")
}

func (failingStore) GetArticle(context.Context, int64) (store.Article, error) {
	return store.Article{}, errors.New("disk on fire!")
}

func (failingStore) ListArticles(context.Context) ([]store.Article, error) {
	return nil, errors.New("disk on fire!")
}

func (failingStore) SearchArticles(context.Context, string, int) ([]store.Article, error) {
	return nil, errors.New("disk on fire!")
}

func (failingStore) GetLatestTriage(context.Context, int64) (store.Triage, error) {
	return store.Triage{}, errors.New("disk on fire!")
}

func (failingStore) CreateTriage(context.Context, store.Triage) (store.Triage, error) {
	return store.Triage{}, errors.New("disk on fire!")
}

func (failingStore) LatestTriages(context.Context) (map[int64]store.Triage, error) {
	return nil, errors.New("disk on fire!")
}

func Test500Errors(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"create ticket", http.MethodPost, "/api/tickets", `{"title": "t", "description": "d"}`},
		{"get ticket", http.MethodGet, "/api/tickets/1", ""},
		{"list tickets", http.MethodGet, "/api/tickets", ""},
		{"get article", http.MethodGet, "/api/kb/1", ""},
		{"list articles", http.MethodGet, "/api/kb", ""},
		{"search articles", http.MethodGet, "/api/kb?q=vpn", ""},
		{"create triage", http.MethodPost, "/api/tickets/1/triage", ""},
		{"get triage", http.MethodGet, "/api/tickets/1/triage", ""},
	}

	mux := newTestServer(failingStore{}, failingStore{}, failingStore{}, nil)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("code = %d, want 500", rec.Code)
			}
		})
	}
}
