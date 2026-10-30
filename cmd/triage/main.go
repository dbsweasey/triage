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
	GetLatestTriage(ctx context.Context, id int64) (store.Triage, error)
	CreateTriage(ctx context.Context, t store.Triage) (store.Triage, error)
}

type triager interface {
	GetResult(ctx context.Context, t store.Ticket, as []store.Article) (triage.Result, error)
}

var (
	_ ticketStore = (*store.MemoryStore)(nil)
	_ ticketStore = (*store.SQLiteStore)(nil)
	_ kbStore     = (*store.SQLiteStore)(nil)
	_ triageStore = (*store.SQLiteStore)(nil)
	_ triager     = (*triage.OpenAIClient)(nil)
)

type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type server struct {
	tickets  ticketStore
	articles kbStore
	triages  triageStore
	triager  triager
}

type createTicketRequest struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	RequesterEmail string `json:"requesterEmail"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "OK", Time: time.Now()})
}

func (s *server) handleCreateTriage(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id. id must be an integer")
		return
	}
	ticket, err := s.tickets.Get(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	articles, err := s.articles.SearchArticles(r.Context(), ticket.Title+" "+ticket.Description, 5)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	result, err := s.triager.GetResult(r.Context(), ticket, articles)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if err := triage.ValidateResult(result, articles); err != nil {
		writeErr(w, r, err)
		return
	}
	tr, err := s.triages.CreateTriage(r.Context(), store.Triage{TicketID: ticket.ID, Priority: result.Priority, Category: result.Category, Summary: result.Summary, Steps: result.Steps, ArticleIDs: result.ArticleIDs})
	if err != nil {
		writeErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, tr)
}

func (s *server) handleGetLatestTriage(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id. id must be an integer")
		return
	}
	ticket, err := s.tickets.Get(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	tr, err := s.triages.GetLatestTriage(r.Context(), ticket.ID)
	if err != nil {
		writeErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, tr)
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

func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	status, msg := http.StatusInternalServerError, "internal server error"

	switch {
	case errors.Is(err, store.ErrNotFound):
		status, msg = http.StatusNotFound, "not found"
	case errors.Is(err, triage.ErrRefused):
		status, msg = http.StatusUnprocessableEntity, "the model declined to triage this ticket"
	case errors.Is(err, triage.ErrUpstream), errors.Is(err, triage.ErrBadResponse), errors.Is(err, triage.ErrIncomplete):
		status, msg = http.StatusBadGateway, "triage service failed"
	}

	if status >= 500 {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}
	writeError(w, status, msg)
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

	mux.HandleFunc("POST /api/tickets/{id}/triage", s.handleCreateTriage)
	mux.HandleFunc("GET /api/tickets/{id}/triage", s.handleGetLatestTriage)

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

	var tri triager = &triage.Fake{}
	if cfg.OpenAIAPIKey != "" {
		tri = triage.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	} else {
		log.Println("OPENAI_API_KEY not set: using fake triager")
	}

	srv := &server{tickets: st, articles: st, triages: st, triager: tri}
	mux := srv.routes()

	log.Printf("listening on %s", address)

	log.Fatal(http.ListenAndServe(address, mux))
}
