package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"triage/internal/store"
)

func (s *server) handleGetArticle(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id. id must be an integer")
		return
	}
	a, err := s.articles.GetArticle(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "article with id "+strconv.FormatInt(id, 10)+" not found.")
		return
	}
	if err != nil {
		log.Printf("get article: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *server) handleListArticles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	// If no query, list articles
	if q == "" {
		list, err := s.articles.ListArticles(r.Context())
		if err != nil {
			log.Printf("list articles: %v", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeJSON(w, http.StatusOK, list)
		return
	}
	// Otherwise, run query
	limitStr := r.URL.Query().Get("limit")
	var limit int
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
	}
	list, err := s.articles.SearchArticles(r.Context(), q, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, list)
}
