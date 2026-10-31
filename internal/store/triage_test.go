package store

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestCreateTriage(t *testing.T) {
	s := createTestDb(t)

	_, err := s.CreateTriage(t.Context(), Triage{TicketID: 1, Priority: "high", Category: "network", Summary: "Restart your VPN", Steps: []string{"1. Restart"}, ArticleIDs: []int64{2}})
	if !strings.Contains(err.Error(), "db insert") {
		t.Fatalf("err = %v, want contains 'db insert'", err)
	}

	ticket := createTicket(t, s, Ticket{Title: "test", Description: "test"})
	got, err := s.CreateTriage(t.Context(), Triage{TicketID: ticket.ID, Priority: "high", Category: "network", Summary: "Restart your VPN", Steps: []string{"1. Restart"}, ArticleIDs: []int64{2}})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want %d", got.ID, 1)
	}
}

func TestGetTriage(t *testing.T) {
	s := createTestDb(t)

	ticket := createTicket(t, s, Ticket{Title: "test", Description: "test"})

	triage := Triage{TicketID: ticket.ID, Priority: "high", Category: "network", Summary: "Restart your VPN", Steps: []string{"1. Restart"}, ArticleIDs: []int64{2}}
	want, err := s.CreateTriage(t.Context(), triage)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	got, err := s.GetLatestTriage(t.Context(), ticket.ID)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if got.ID != want.ID {
		t.Errorf("id = %d, want %d", got.ID, want.ID)
	}
	if got.Category != want.Category {
		t.Errorf("category = %v, want %v", got.Category, want.Category)
	}
	if got.Priority != want.Priority {
		t.Errorf("priority = %v, want %v", got.Priority, want.Priority)
	}
	if got.Summary != want.Summary {
		t.Errorf("summary = %v, want %v", got.Summary, want.Summary)
	}
	if !slices.Equal(got.ArticleIDs, want.ArticleIDs) {
		t.Errorf("articleIds = %v, want %v", got.ArticleIDs, want.ArticleIDs)
	}
	if !slices.Equal(got.Steps, want.Steps) {
		t.Errorf("steps = %v, want %v", got.Steps, want.Steps)
	}

	// Add a couple triages, ensure get pulls most recent
	triage.Summary = "summary2"
	s.CreateTriage(t.Context(), triage)
	triage.Summary = "summary3"
	want, _ = s.CreateTriage(t.Context(), triage)

	got, err = s.GetLatestTriage(t.Context(), want.TicketID)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %d, want %d", got.ID, want.ID)
	}
	if got.Summary != want.Summary {
		t.Errorf("summary = %q, want %q", got.Summary, want.Summary)
	}

	// No triage with given ticket ID
	_, err = s.GetLatestTriage(t.Context(), 2)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want %v", err, ErrNotFound)
	}
}

func TestLatestTriages(t *testing.T) {
	s := createTestDb(t)

	ticket := createTicket(t, s, Ticket{Title: "test", Description: "test"})
	createTicket(t, s, Ticket{Title: "test2", Description: "test2"})

	_, err := s.CreateTriage(t.Context(), Triage{TicketID: ticket.ID, Priority: "high", Category: "network", Summary: "Restart your VPN", Steps: []string{"1. Restart"}, ArticleIDs: []int64{}})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	want, err := s.CreateTriage(t.Context(), Triage{TicketID: ticket.ID, Priority: "medium", Category: "software", Summary: "VPN not working", Steps: []string{"1. Restart", "2. Try again"}, ArticleIDs: []int64{}})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	triages, err := s.LatestTriages(t.Context())
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if len(triages) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(triages))
	}
	got, ok := triages[want.TicketID]
	if !ok {
		t.Fatalf("ok = false, want true")
	}
	if got.ID != want.ID {
		t.Errorf("id = %d, want %d", got.ID, want.ID)
	}
}
