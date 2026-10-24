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
	serverMux.HandleFunc("/healthz", readinessHandler)
	serverMux.HandleFunc("/metrics", apiCfg.metricsHandler)
	serverMux.HandleFunc("/reset", apiCfg.resetHandler)
	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Fail running the server")
	}
}

func readinessHandler(respWriter http.ResponseWriter, req *http.Request) {
	respWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	respWriter.WriteHeader(http.StatusOK)
	_, err := respWriter.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Printf("fail writing response: %v", err)
	}
}

func (cfg *apiConfig) metricsHandler(respWriter http.ResponseWriter, req *http.Request) {
	respWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	respWriter.WriteHeader(http.StatusOK)
	countStr := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
	_, err := respWriter.Write([]byte(countStr))
	if err != nil {
		log.Printf("fail writing response: %v", err)
	}
}

func (cfg *apiConfig) resetHandler(respWriter http.ResponseWriter, req *http.Request) {
	respWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	respWriter.WriteHeader(http.StatusOK)
	count := cfg.fileserverHits.Load()
	cfg.fileserverHits.Add(-count)
	_, err := respWriter.Write([]byte("fileserverHits count reseted."))
	if err != nil {
		log.Printf("fail writing response: %v", err)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		handler.ServeHTTP(w, req)
	})
}
