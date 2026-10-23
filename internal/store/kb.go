package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed seed/kb.json
var seedKB []byte

func seedArticles(ctx context.Context, s *SQLiteStore) error {

	var articles []Article
	if err := json.Unmarshal(seedKB, &articles); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	defer tx.Rollback()

	q := `
		INSERT INTO kb_articles (title, body, tags)
		VALUES (?, ?, ?)
	`
	for _, a := range articles {
		if _, err := tx.ExecContext(ctx, q, a.Title, a.Body, strings.Join(a.Tags, ",")); err != nil {
			return fmt.Errorf("seed article %q: %w", a.Title, err)
		}
	}
	return tx.Commit()
}
