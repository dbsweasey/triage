package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"triage/internal/config"
)

type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := healthResponse{Status: "ok", Time: time.Now()}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("write health response: %v", err)
	}
}

func main() {
	name := "Triage"
	fmt.Printf("%s starting...\n", name)

	cfg := config.Load()
	address := "localhost:" + cfg.Port

	log.Printf("openai key set: %t", cfg.OpenAIAPIKey != "")

	log.Printf("listening on %s", address)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	log.Fatal(http.ListenAndServe(address, mux))
}
