package store

import (
	"cmp"
	"errors"
	"maps"
	"slices"
	"sync"
	"time"
)

// Ticket stores the data related to a ticket
type Ticket struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	RequesterEmail string    `json:"requesterEmail"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

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
func (s *MemoryStore) Create(t Ticket) Ticket {
	// Lock before accessing map
	s.mu.Lock()
	defer s.mu.Unlock()

	t.CreatedAt = time.Now()
	t.Status = "open"
	t.ID = s.nextID

	s.tickets[s.nextID] = t

	s.nextID++

	return t
}

// Get returns the ticket with the ID, error otherwise
func (s *MemoryStore) Get(id int64) (Ticket, error) {
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
func (s *MemoryStore) List() []Ticket {
	// Lock before accessing map
	s.mu.Lock()
	defer s.mu.Unlock()

	list := slices.SortedFunc(maps.Values(s.tickets), func(a, b Ticket) int {
		return cmp.Compare(b.ID, a.ID)
	})
	if list == nil {
		return []Ticket{}
	}
	return list
}

// ErrNotFound acts as a sentinel error for missing ticket
var ErrNotFound = errors.New("ticket not found")
