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
	SecretJWT      string
}

func main() {
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	dbQueries := database.New(db)
	if err != nil {
		log.Fatalf("Fail opening db: %v", err)
	}
	apiCfg := apiConfig{
		DbQueries: dbQueries,
		SecretJWT: os.Getenv("JWT_SECRET"),
	}
	fmt.Printf("Hello, World!")
	mux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(appHandler))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("POST /api/login", apiCfg.loginHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	mux.HandleFunc("GET /api/users", apiCfg.getUsers)
	mux.HandleFunc("GET /api/users/{userID}", apiCfg.getUserByIDHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirpHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpByIDHandler)
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Fail running the server")
	}
}
