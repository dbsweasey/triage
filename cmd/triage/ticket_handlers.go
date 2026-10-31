package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"triage/internal/store"
)

type ticketListItem struct {
	store.Ticket
	Triage *store.Triage `json:"triage"`
}

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

func (s *server) handleListTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := s.tickets.List(r.Context())
	if err != nil {
		log.Printf("list tickets: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	latest, err := s.triages.LatestTriages(r.Context())
	if err != nil {
		log.Printf("list tickets: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	items := make([]ticketListItem, 0, len(tickets))
	for _, t := range tickets {
		item := ticketListItem{Ticket: t}
		if tr, ok := latest[t.ID]; ok {
			item.Triage = &tr
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, items)
}
