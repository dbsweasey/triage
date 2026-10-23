package store

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// storeUnderTest is the behavior every ticket store must have
type storeUnderTest interface {
	Create(ctx context.Context, t Ticket) (Ticket, error)
	Get(ctx context.Context, id int64) (Ticket, error)
	List(ctx context.Context) ([]Ticket, error)
}

func TestMemoryStore(t *testing.T) {
	runStoreTests(t, func(t *testing.T) storeUnderTest { return NewMemoryStore() })
}

func TestSQLiteStore(t *testing.T) {
	runStoreTests(t, func(t *testing.T) storeUnderTest { return createTestDb(t) })
}

func createTicket(t *testing.T, s storeUnderTest, ticket Ticket) Ticket {
	t.Helper()
	ticket, err := s.Create(t.Context(), ticket)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return ticket
}

// runStoreTests runs shared behavior tests
func runStoreTests(t *testing.T, newStore func(t *testing.T) storeUnderTest) {
	t.Run("Create", func(t *testing.T) {
		s := newStore(t)

		for want := int64(1); want <= 3; want++ {
			got, err := s.Create(t.Context(), Ticket{Title: "x"})
			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if got.ID != want {
				t.Errorf("ID = %d, want %d", got.ID, want)
			}
			if got.Status != "open" {
				t.Errorf("status = %v, want open", got.Status)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		s := newStore(t)

		want := createTicket(t, s, Ticket{Title: "test", Description: "d", RequesterEmail: "a@b.com"})

		got, err := s.Get(t.Context(), want.ID)
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if got.ID != want.ID {
			t.Errorf("ID = %d, want %d", got.ID, want.ID)
		}
		if got.Title != want.Title {
			t.Errorf("Title = %v, want %v", got.Title, want.Title)
		}
		if got.Description != want.Description {
			t.Errorf("Description = %v, want %v", got.Description, want.Description)
		}
		if got.RequesterEmail != want.RequesterEmail {
			t.Errorf("RequesterEmail = %v, want %v", got.RequesterEmail, want.RequesterEmail)
		}
		if got.Status != want.Status {
			t.Errorf("Status = %v, want %v", got.Status, want.Status)
		}
		if !got.CreatedAt.Equal(want.CreatedAt) {
			t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, want.CreatedAt)
		}

		_, err = s.Get(t.Context(), 2)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("err = %v, want %v", err, ErrNotFound)
		}
	})

	t.Run("List", func(t *testing.T) {
		s := newStore(t)

		got, err := s.List(t.Context())
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if got == nil {
			t.Fatalf("list = nil, want []")
		}
		if len(got) != 0 {
			t.Errorf("length = %d, want 0", len(got))
		}

		createTicket(t, s, Ticket{Title: "x"})
		createTicket(t, s, Ticket{Title: "y"})
		createTicket(t, s, Ticket{Title: "z1"})

		got, err = s.List(t.Context())
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if len(got) != 3 {
			t.Fatalf("len = %d, want 3", len(got))
		}
		if got[0].ID != 3 {
			t.Errorf("ID = %d, want 3", got[0].ID)
		}
	})

	t.Run("Concurrency", func(t *testing.T) {
		s := newStore(t)

		var wg sync.WaitGroup
		for range 100 {
			wg.Go(func() {
				if _, err := s.Create(t.Context(), Ticket{Title: "x"}); err != nil {
					t.Errorf("create: %v", err)
				}
			})
		}
		wg.Wait()

		got, err := s.List(t.Context())
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if len(got) != 100 {
			t.Errorf("len = %d, want 100", len(got))
		}
	})
}
