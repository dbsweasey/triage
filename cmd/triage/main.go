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
	"triage/internal/triage"
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

type triageStore interface {
	GetTriage(ctx context.Context, id int64) (store.Triage, error)
	CreateTriage(ctx context.Context, t store.Triage) (store.Triage, error)
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
	triages  triageStore
	llm      *triage.OpenAIClient
}

type createTicketRequest struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	RequesterEmail string `json:"requesterEmail"`
}

type createTriageRequest struct {
	TicketId int64 `json:"ticketId"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "OK", Time: time.Now()})
}

func (s *server) handleTriage(w http.ResponseWriter, r *http.Request) {
	var input createTriageRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}
	ticket, err := s.tickets.Get(r.Context(), input.TicketId)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ticket with id "+strconv.FormatInt(input.TicketId, 10)+" not found.")
		return
	}
	if err != nil {
		log.Printf("get ticket: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	articles, err := s.articles.SearchArticles(r.Context(), ticket.Title+" "+ticket.Description, 5)
	if err != nil {
		log.Printf("search articles: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	result, err := s.llm.GetResult(r.Context(), ticket, articles)
	if err != nil {
		log.Printf("call llm: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if err := triage.ValidateResult(result, articles); err != nil {
		log.Printf("validate result: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	triage, err := s.triages.CreateTriage(r.Context(), store.Triage{TicketID: ticket.ID, Priority: result.Priority, Category: result.Category, Summary: result.Summary, Steps: result.Steps, ArticleIDs: result.ArticleIDs})

	writeJSON(w, http.StatusCreated, triage)
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
	mux.HandleFunc("GET /api/kb", s.handleListArticles)

	mux.HandleFunc("POST /api/triage", s.handleTriage)

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

	llm := triage.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)

	srv := &server{tickets: st, articles: st, triages: st, llm: llm}
	mux := srv.routes()

	log.Printf("listening on %s", address)

	log.Fatal(http.ListenAndServe(address, mux))
}
