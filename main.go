package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Muhammad-Romli/Chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	FileServerHits atomic.Int32
	DbQueries      *database.Queries
}

func main() {
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	dbQueries := database.New(db)
	if err != nil {
		log.Fatalf("Fail opening db: %w", err)
	}
	apiCfg := apiConfig{DbQueries: dbQueries}
	fmt.Printf("Hello, World!")
	mux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(appHandler))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getUsersHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getUserByIDHandler)
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Fail running the server")
	}
}
