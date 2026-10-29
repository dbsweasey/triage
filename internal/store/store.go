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

type Triage struct {
	ID         int64     `json:"id"`
	TicketID   int64     `json:"ticketId"`
	Priority   string    `json:"priority"`
	Category   string    `json:"category"`
	Summary    string    `json:"summary"`
	Steps      []string  `json:"steps"`
	ArticleIDs []int64   `json:"articleIds"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ErrNotFound acts as a sentinel error for missing item
var ErrNotFound = errors.New("not found")
