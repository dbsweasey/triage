package store

import (
	"path/filepath"
	"testing"
)

// createTestDb creates a temporary DB within a temp directory for testing SQLiteStore
func createTestDb(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := OpenSQLite(t.Context(), dbPath)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestSQLitePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	s1, err := OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}

	want := createTicket(t, s1, Ticket{Title: "test", Description: "d", RequesterEmail: "a@b.com"})
	s1.Close()

	s2, err := OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}

	got, err := s2.Get(t.Context(), 1)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %d, want %d", got.ID, want.ID)
	}
	if got.Title != want.Title {
		t.Errorf("Title = %q, want %q", got.Title, want.Title)
	}
	if got.Description != want.Description {
		t.Errorf("Description = %q, want %q", got.Description, want.Description)
	}

	t.Cleanup(func() { s2.Close() })
}
