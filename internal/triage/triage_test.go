package triage

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"triage/internal/store"
)

func TestBuildPrompt(t *testing.T) {
	ticket := store.Ticket{Title: "New ticket", Description: "This ticket is new"}
	sys, user := buildPrompt(
		ticket,
		[]store.Article{
			{ID: 1, Title: "Test article", Body: "This article is a test", Tags: []string{"test", "article"}},
			{ID: 2, Title: "Another article", Body: "Another test article", Tags: []string{"another", "article"}},
		})
	if !strings.Contains(sys, "# Role") {
		t.Errorf("contains '# Role' = false, want true")
	}
	if !strings.Contains(sys, "# Input") {
		t.Errorf("contains '# Input' = false, want true")
	}
	if !strings.Contains(sys, "# Goal") {
		t.Errorf("contains '# Goal' = false, want true")
	}
	want :=
		`Help me resolve the following ticket.
Title: New ticket
Description: This ticket is new

Here are the relevant knowledge articles:

Article #1 - Test article
	This article is a test
	Tags: [test article]

Article #2 - Another article
	Another test article
	Tags: [another article]
`
	if user != want {
		t.Errorf("user = \n%v\nwant\n%v\nDONE.", user, want)
	}

	_, user = buildPrompt(ticket, []store.Article{})
	want = `Help me resolve the following ticket.
Title: New ticket
Description: This ticket is new`
	if user != want {
		t.Errorf("user = \n%v\nwant\n%v\nDONE.", user, want)
	}
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

func TestValidateResponse(t *testing.T) {
	tests := []struct {
		name       string
		result     Result
		errWant    error
		errContain string
	}{
		{"invalid priority", buildResponse("priority"), ErrBadResponse, "invalid priority"},
		{"good response", buildResponse("good"), nil, ""},
		{"invalid steps", buildResponse("steps"), ErrBadResponse, "no steps"},
		{"invalid category", buildResponse("category"), ErrBadResponse, "invalid category"},
		{"invalid articles", buildResponse("articles"), ErrBadResponse, "invalid article"},
	}
	articles := []store.Article{{ID: 1, Title: "Test", Body: "test", Tags: []string{}}, {ID: 2, Title: "Test", Body: "test", Tags: []string{}}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateResult(tc.result, articles)
			if tc.errWant == nil {
				if got != nil {
					t.Fatalf("err = %v, want nil", got)
				}
				return
			}
			if !errors.Is(got, tc.errWant) {
				t.Errorf("err = %v, want %v", got, tc.errWant)
			}
			if !strings.Contains(got.Error(), tc.errContain) {
				t.Errorf("err = %v, want contains %v", got, tc.errContain)
			}
		})
	}
}

func buildResponse(issue string) Result {
	r := Result{Priority: "high", Category: "network", Summary: "test", Steps: []string{"1. test", "2. test"}, ArticleIDs: []int64{1, 2}}
	if issue == "priority" {
		r.Priority = "womp"
	}
	if issue == "category" {
		r.Category = "womp"
	}
	if issue == "steps" {
		r.Steps = []string{}
	}
	if issue == "articles" {
		r.ArticleIDs = []int64{1, 67, 69}
	}
	return r
}

func messageResponse(status string, part string) string {
	return `{"status":"` + status + `","output":[{"type":"message","content":[` + part + `]}]}`
}
