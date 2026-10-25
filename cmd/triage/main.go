package main

import (
	"context"
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

// ticketStore interface defines signatures for interacting with the ticketStore
type ticketStore interface {
	Create(ctx context.Context, t store.Ticket) (store.Ticket, error)
	Get(ctx context.Context, id int64) (store.Ticket, error)
	List(ctx context.Context) ([]store.Ticket, error)
}

type kbStore interface {
	GetArticle(ctx context.Context, id int64) (store.Article, error)
	ListArticles(ctx context.Context) ([]store.Article, error)
	SearchArticles(ctx context.Context, text string, limit int) ([]store.Article, error)
}

var (
	_ ticketStore = (*store.MemoryStore)(nil)
	_ ticketStore = (*store.SQLiteStore)(nil)
	_ kbStore     = (*store.SQLiteStore)(nil)
)

type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type server struct {
	tickets  ticketStore
	articles kbStore
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
	list, err := s.tickets.List(r.Context())
	if err != nil {
		log.Printf("list tickets: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

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

func parseID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	return id, err
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("POST /api/tickets", s.handleCreateTicket)
	mux.HandleFunc("GET /api/tickets/{id}", s.handleGetTicket)
	mux.HandleFunc("GET /api/tickets", s.handleListTickets)

	mux.HandleFunc("GET /api/kb/{id}", s.handleGetArticle)

	return mux
}

func main() {
	name := "Triage"
	fmt.Printf("%s starting...\n", name)

	cfg := config.Load()
	address := "localhost:" + cfg.Port

	log.Printf("openai key set: %t", cfg.OpenAIAPIKey != "")

	st, err := store.OpenSQLite(context.Background(), cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	defer st.Close()

	srv := &server{tickets: st, articles: st}
	mux := srv.routes()

	log.Printf("listening on %s", address)

	log.Fatal(http.ListenAndServe(address, mux))
}
