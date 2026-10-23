package store

import (
	"errors"
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

type Article struct {
	ID    int64    `json:"id"`
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Tags  []string `json:"tags"`
}

// ErrNotFound acts as a sentinel error for missing ticket
var ErrNotFound = errors.New("not found")
