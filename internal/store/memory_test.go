package store

import (
	"errors"
	"sync"
	"testing"
)

func TestCreate(t *testing.T) {
	s := NewMemoryStore()

	for want := int64(1); want <= 3; want++ {
		got := s.Create(Ticket{Title: "x"})
		if got.ID != want {
			t.Errorf("ID = %d, want %d", got.ID, want)
		}
		if got.Status != "open" {
			t.Errorf("status = %v, want open", got.Status)
		}
	}
}

func TestGet(t *testing.T) {
	s := NewMemoryStore()

	s.Create(Ticket{Title: "test"})

	ticket, err := s.Get(1)
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if ticket.ID != 1 {
		t.Errorf("ID = %d, want 1", ticket.ID)
	}

	_, err = s.Get(2)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want %v", err, ErrNotFound)
	}
}

func TestList(t *testing.T) {
	s := NewMemoryStore()

	got := s.List()
	if got == nil {
		t.Fatalf("list = nil, want []")
	}
	if len(got) != 0 {
		t.Errorf("length = %d, want 0", len(got))
	}

	s.Create(Ticket{Title: "y"})
	s.Create(Ticket{Title: "x"})
	s.Create(Ticket{Title: "w"})

	got = s.List()
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].ID != 3 {
		t.Errorf("ID = %d, want 3", got[0].ID)
	}
}

func TestConcurrency(t *testing.T) {
	s := NewMemoryStore()

	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() { s.Create(Ticket{Title: "x"}) })
	}
	wg.Wait()

	if got := len(s.List()); got != 100 {
		t.Errorf("len = %d, want 100", got)
	}
}
