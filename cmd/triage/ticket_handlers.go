package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"triage/internal/store"
)

func (s *server) handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	var input createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid schema for new ticket")
		return
	}
	if input.Description == "" || input.Title == "" {
		writeError(w, http.StatusBadRequest, "title and description are required")
		return
	}
	t := store.Ticket{Title: input.Title, Description: input.Description, RequesterEmail: input.RequesterEmail}
	t, err := s.tickets.Create(r.Context(), t)

	if err != nil {
		log.Printf("create ticket: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (s *server) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id. id must be an integer")
		return
	}
	t, err := s.tickets.Get(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ticket with id "+strconv.FormatInt(id, 10)+" not found.")
		return
	}
	if err != nil {
		log.Printf("get ticket: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, t)
}
