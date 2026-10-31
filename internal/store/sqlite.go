package store

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore holds pointer to the open db
type SQLiteStore struct {
	db *sql.DB
}

type ticketListItem struct {
	Ticket
	Triage *Triage `json:"triage"`
}

//go:embed schema.sql
var schema string

// scanner defines an interface for types that implement Scan
type scanner interface {
	Scan(dest ...any) error
}

// scanTicket runs Scan on types that impliment scanner. Pulls SQL data into a Ticket
func scanTicket(s scanner) (Ticket, error) {
	var t Ticket
	err := s.Scan(&t.ID, &t.Title, &t.Description, &t.RequesterEmail, &t.Status, &t.CreatedAt)
	return t, err
}

// OpenSQLite opens the the sqlite DB and returns it in an SQLiteStore
func OpenSQLite(ctx context.Context, path string) (*SQLiteStore, error) {
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("db ping context: %w", err)
	}

	if _, err = db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("db exec context: %w", err)
	}

	s := SQLiteStore{db: db}
	err = s.seedArticles(ctx)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("seed: %w", err)
	}

	return &s, nil
}

// Close the db in SQLiteStore
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Create creates a new ticket and adds it to the db
func (s *SQLiteStore) Create(ctx context.Context, t Ticket) (Ticket, error) {
	t.CreatedAt = time.Now().UTC()
	t.Status = "open"

	q := `
		INSERT INTO tickets (title, description, requester_email, status, created_at)
		VALUES (?, ?, ?, ?, ?)`

	res, err := s.db.ExecContext(ctx, q, t.Title, t.Description, t.RequesterEmail, t.Status, t.CreatedAt)
	if err != nil {
		return Ticket{}, fmt.Errorf("db insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Ticket{}, fmt.Errorf("db insert: %w", err)
	}

	t.ID = id

	return t, nil
}

// Get returns the ticket with the ID, error otherwise
func (s *SQLiteStore) Get(ctx context.Context, id int64) (Ticket, error) {
	q := `
		SELECT id, title, description, requester_email, status, created_at
		FROM tickets
		WHERE id = ?`

	row := s.db.QueryRowContext(ctx, q, id)
	t, err := scanTicket(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Ticket{}, ErrNotFound
	}
	if err != nil {
		return Ticket{}, fmt.Errorf("get ticket %d: %w", id, err)
	}

	return t, nil
}

// List retrieves all tickets, newest first
func (s *SQLiteStore) List(ctx context.Context) ([]Ticket, error) {
	list := []Ticket{}
	q := `
		SELECT id, title, description, requester_email, status, created_at
		FROM tickets
		ORDER BY id DESC`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("list tickets: %w", err)
		}
		list = append(list, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}

	return list, nil
}

func (s *SQLiteStore) UpdateStatus(ctx context.Context, id int64, status string) (Ticket, error) {
	// TODO implement
	return Ticket{}, nil
}
