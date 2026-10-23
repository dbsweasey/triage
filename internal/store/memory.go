package store

import (
	"cmp"
	"context"
	"maps"
	"slices"
	"sync"
	"time"
)

// MemoryStore stores tickets in memory. Mutex lock to avoid race conditions
type MemoryStore struct {
	mu      sync.Mutex
	tickets map[int64]Ticket
	nextID  int64 // Next available ticket ID
}

// NewMemoryStore constructor for in-memory ticket storage
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tickets: make(map[int64]Ticket), nextID: 1}
}

// Create updates MemoryStore with new ticket
func (s *MemoryStore) Create(ctx context.Context, t Ticket) (Ticket, error) {
	// Lock before accessing map
	s.mu.Lock()
	defer s.mu.Unlock()

	t.CreatedAt = time.Now()
	t.Status = "open"
	t.ID = s.nextID

	s.tickets[s.nextID] = t

	s.nextID++

	return t, nil
}

// Get returns the ticket with the ID, error otherwise
func (s *MemoryStore) Get(ctx context.Context, id int64) (Ticket, error) {
	// Lock before accessing map
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tickets[id]

	if !ok {
		return Ticket{}, ErrNotFound
	}
	return t, nil
}

// List retrieves all tickets, newest first
func (s *MemoryStore) List(ctx context.Context) ([]Ticket, error) {
	// Lock before accessing map
	s.mu.Lock()
	defer s.mu.Unlock()

	list := slices.SortedFunc(maps.Values(s.tickets), func(a, b Ticket) int {
		return cmp.Compare(b.ID, a.ID)
	})
	if list == nil {
		return []Ticket{}, nil
	}
	return list, nil
}
