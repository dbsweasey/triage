package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func scanTriage(s scanner) (Triage, error) {
	var t Triage
	var steps []byte
	var articles []byte
	err := s.Scan(&t.ID, &t.TicketID, &t.Priority, &t.Category, &t.Summary, &steps, &articles, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Triage{}, ErrNotFound
	}
	if err != nil {
		return Triage{}, fmt.Errorf("get triage: %w", err)
	}
	err = json.Unmarshal(steps, &t.Steps)
	if err != nil {
		return Triage{}, fmt.Errorf("unmarshal steps: %w", err)
	}
	err = json.Unmarshal(articles, &t.ArticleIDs)
	if err != nil {
		return Triage{}, fmt.Errorf("unmarshal articles: %w", err)
	}
	return t, err
}

// CreateTriage saves a triage result to the store
func (s *SQLiteStore) CreateTriage(ctx context.Context, t Triage) (Triage, error) {
	q := `
		INSERT INTO triages (ticket_id, priority, category, summary, steps, article_ids, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	steps, err := json.Marshal(t.Steps)
	if err != nil {
		return Triage{}, fmt.Errorf("save triage: %w", err)
	}
	articles, err := json.Marshal(t.ArticleIDs)
	if err != nil {
		return Triage{}, fmt.Errorf("save triage: %w", err)
	}

	t.CreatedAt = time.Now().UTC()

	res, err := s.db.ExecContext(ctx, q, t.TicketID, t.Priority, t.Category, t.Summary, steps, articles, t.CreatedAt)
	if err != nil {
		return Triage{}, fmt.Errorf("db insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Triage{}, fmt.Errorf("db insert: %w", err)
	}
	t.ID = id

	return t, nil
}

// GetLatestTriage retrieves a triage based on the ticket ID
func (s *SQLiteStore) GetLatestTriage(ctx context.Context, ticketID int64) (Triage, error) {
	q := `
		SELECT id, ticket_id, priority, category, summary, steps, article_ids, created_at
		FROM triages
		WHERE ticket_id = ?
		ORDER BY id DESC
		LIMIT 1
	`

	res := s.db.QueryRowContext(ctx, q, ticketID)
	t, err := scanTriage(res)

	return t, err
}

// LatestTriages returns the most recent triage for each ticket, keyed by ticket ID.
// Tickets that have never been triaged are not included.
func (s *SQLiteStore) LatestTriages(ctx context.Context) (map[int64]Triage, error) {
	q := `
		SELECT id, ticket_id, priority, category, summary, steps, article_ids, created_at
		FROM triages
		WHERE id IN (SELECT MAX(id) FROM triages GROUP BY ticket_id)
	`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("latest triages: %w", err)
	}
	defer rows.Close()

	triages := make(map[int64]Triage)

	for rows.Next() {
		t, err := scanTriage(rows)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		triages[t.TicketID] = t
	}
	return triages, nil
}
