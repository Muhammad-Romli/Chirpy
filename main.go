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
	serverMux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(appHandler))
	serverMux.HandleFunc("GET /api/healthz", readinessHandler)
	serverMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	serverMux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	serverMux.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Fail running the server")
	}
}
