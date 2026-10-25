package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	apiCfg := apiConfig{}
	fmt.Printf("Hello, World!")
	serverMux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(appHandler))
	serverMux.HandleFunc("GET /api/healthz", readinessHandler)
	serverMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	serverMux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Fail running the server")
	}
}
