package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"triage/internal/store"
)

func newTestServer() http.Handler {
	srv := &server{tickets: store.NewMemoryStore()}
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

	mux := newTestServer()

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

	mux := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/api/tickets", strings.NewReader(`{"title": "test", "description": "test"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup: create status = %d, want 201", rec.Code)
	}

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
