package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"triage/internal/config"
	"triage/internal/store"
)

type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type server struct {
	tickets *store.MemoryStore
}

type createTicketRequest struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	RequesterEmail string `json:"requesterEmail"`
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
	t = s.tickets.Create(t)

	writeJSON(w, http.StatusCreated, t)
}

func (s *server) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id. id must be an integer")
		return
	}
	t, err := s.tickets.Get(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ticket with id "+idStr+" not found.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *server) handleGetTickets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.tickets.List())
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "OK", Time: time.Now()})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("POST /api/tickets", s.handleCreateTicket)
	mux.HandleFunc("GET /api/tickets/{id}", s.handleGetTicket)
	mux.HandleFunc("GET /api/tickets", s.handleGetTickets)

	return mux
}

func main() {
	name := "Triage"
	fmt.Printf("%s starting...\n", name)

	cfg := config.Load()
	address := "localhost:" + cfg.Port

	log.Printf("openai key set: %t", cfg.OpenAIAPIKey != "")

	srv := &server{tickets: store.NewMemoryStore()}
	mux := srv.routes()

	log.Printf("listening on %s", address)

	log.Fatal(http.ListenAndServe(address, mux))
}
