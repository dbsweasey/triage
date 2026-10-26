package triage

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"triage/internal/store"
)

func TestBuildPrompt(t *testing.T) {
	sys, user := buildPrompt(
		store.Ticket{Title: "New ticket", Description: "This ticket is new"},
		[]store.Article{
			{Title: "Test article", Body: "This article is a test", Tags: []string{"test", "article"}},
			{Title: "Another article", Body: "Another test article", Tags: []string{"another", "article"}},
		})
	t.Log(sys)
	t.Log(user)
}

func TestGetResultHappy(t *testing.T) {
	var gotReq requestBody
	var gotAuth string
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		json.NewDecoder(r.Body).Decode(&gotReq)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "completed",
			"output": [{
				"type": "message",
				"content": [{
					"type": "output_text",
					"text": "{\"priority\":\"high\",\"category\":\"network\",\"summary\":\"Uninstall and reinstall VPN\",\"steps\":[\"1. Uninstall\",\"2. Reinstall\"],\"articleIds\":[1]}"
				}]
			}]
		}`))
	}))

	c := NewOpenAIClient("test-key", "fake-model")
	c.client = srv.Client()

	res, err := c.GetResult(t.Context(), store.Ticket{Title: "VPN is broken", Description: "My VPN fails to connect"}, []store.Article{{Title: "VPN issues", Body: "Uninstall and reinstall is the simple fix to VPN issues"}})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("auth = %v, want Bearer test-key", gotAuth)
	}
	if res.Priority != "high" {
		t.Errorf("priority = %v, want high", res.Priority)
	}
	if res.Category != "network" {
		t.Errorf("category = %v, want network", res.Category)
	}
	if len(res.Steps) != 2 {
		t.Errorf("len(steps) = %v, want 2", len(res.Steps))
	}
	if !gotReq.Text.Format.Strict {
		t.Errorf("text/format/strict = false, want true")
	}
	if len(res.ArticleIDs) != 1 {
		t.Errorf("len(articleIds) = %d, want 1", len(res.ArticleIDs))
	}
}

func TestGetResultBad(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantError error
	}{
		{"server error", http.StatusInternalServerError, `{"error": {"message":"internal server error"}}`, ErrUpstream},
		{"incomplete", http.StatusOK, messageResponse("incomplete", ``), ErrIncomplete},
		{"refusal", http.StatusOK, messageResponse("completed", `{"type":"refusal","refusal":"I can't help with that"}`), ErrRefused},
		{"invalid inner JSON", http.StatusOK, messageResponse("completed", `{"type":"output_text}`), ErrBadResponse},
		{"no outer_text", http.StatusOK, messageResponse("completed", `{"type":"message": "message":"hello"}`), ErrBadResponse},
		{"invalid outer JSON", http.StatusOK, `{"status": "completed}`, ErrBadResponse},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			c := NewOpenAIClient("test-key", "fake-model")
			c.client = srv.Client()

			_, err := c.GetResult(t.Context(), store.Ticket{Title: "VPN is broken", Description: "My VPN fails to connect"}, []store.Article{{Title: "VPN issues", Body: "Uninstall and reinstall is the simple fix to VPN issues"}})
			if !errors.Is(err, tc.wantError) {
				t.Errorf("err = %v, want %v", err, tc.wantError)
			}

		})
	}
}

func messageResponse(status string, part string) string {
	return `{"status":"` + status + `","output":[{"type":"message","content":[` + part + `]}]}`
}
