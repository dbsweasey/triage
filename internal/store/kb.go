package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

//go:embed seed/kb.json
var seedKB []byte

var stopwords = map[string]struct{}{
	"the": {}, "a": {}, "is": {}, "an": {}, "and": {}, "it": {},
	"in": {}, "on": {}, "but": {}, "by": {}, "for": {}, "me": {},
	"my": {}, "we": {}, "you": {}, "he": {}, "she": {}, "they": {},
	"are": {}, "was": {}, "have": {}, "do": {}, "can": {}, "will": {},
	"am": {}, "or": {},
}

func parseTags(t string) []string {
	if t == "" {
		return []string{}
	}
	tags := strings.Split(t, ",")
	return tags
}

func scanArticle(s scanner) (Article, error) {
	var a Article
	var tags string
	err := s.Scan(&a.ID, &a.Title, &a.Body, &tags)
	a.Tags = parseTags(tags)
	return a, err
}

func (s *SQLiteStore) seedArticles(ctx context.Context) error {

	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM kb_articles`).Scan(&count); err != nil {
		return fmt.Errorf("count: %w", err)
	}
	if count != 0 {
		return nil
	}

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

func (s *SQLiteStore) GetArticle(ctx context.Context, id int64) (Article, error) {
	q := `
		SELECT id, title, body, tags
		FROM kb_articles
		WHERE id = ?
	`
	row := s.db.QueryRowContext(ctx, q, id)
	a, err := scanArticle(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Article{}, ErrNotFound
	}
	if err != nil {
		return Article{}, fmt.Errorf("get article %d: %w", id, err)
	}

	return a, nil
}

func (s *SQLiteStore) ListArticles(ctx context.Context) ([]Article, error) {
	list := []Article{}
	q := `
		SELECT id, title, body, tags
		FROM kb_articles
		ORDER BY title
	`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, fmt.Errorf("list articles: %w", err)
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}

	return list, nil
}

func (s *SQLiteStore) SearchArticles(ctx context.Context, text string, limit int) ([]Article, error) {
	query := buildFTSQuery(text)
	if query == "" {
		return []Article{}, nil
	}
	if limit <= 0 {
		limit = 5
	}
	list := []Article{}
	q := `
		SELECT a.id, a.title, a.body, a.tags
		FROM kb_fts
		JOIN kb_articles a ON a.id = kb_fts.rowid
		WHERE kb_fts MATCH ?
		ORDER BY bm25(kb_fts, 10.0, 1.0, 5.0)
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, q, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search articles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, fmt.Errorf("search articles: %w", err)
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search articles: %w", err)
	}

	return list, nil
}

// buildFTSQuery filters stopwords and constructs a string to use in fts5
func buildFTSQuery(text string) string {
	text = strings.Replace(text, "'", "", -1)

	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsDigit(r) && !unicode.IsLetter(r)
	})

	filtered := []string{}
	for _, w := range words {
		if _, ok := stopwords[strings.ToLower(w)]; ok {
			continue
		}
		if len(w) < 2 {
			continue
		}
		filtered = append(filtered, `"`+strings.ToLower(w)+`"`)
	}
	return strings.Join(filtered, " OR ")
}
