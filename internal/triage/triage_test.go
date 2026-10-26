package triage

import (
	"encoding/json"
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

func TestGetResult(t *testing.T) {
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
}
