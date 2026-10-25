package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

type ReqParams struct {
	Body string `json:"body"`
}

type RespParams struct {
	Error string `json:"error"`
	Valid bool   `json:"valid"`
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

func readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Fail writing response: %v", err)
	}
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	countStr := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
	_, err := w.Write([]byte(countStr))
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Fail writing response: %v", err)
	}
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	count := cfg.fileserverHits.Load()
	cfg.fileserverHits.Add(-count)
	_, err := w.Write([]byte("fileserverHits count reseted."))
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Fail writing response: %v", err)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		handler.ServeHTTP(w, req)
	})
}

func validateChirpHandler(w http.ResponseWriter, req *http.Request) {
	params := ReqParams{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	if len(params.Body) > 140 {
		errMsg := "Chirp is too long"
		respError(w, errMsg, 400)
		return
	}
	resp := RespParams{Valid: true}
	respJSON(w, 200, resp)
}

func respError(w http.ResponseWriter, errMsg string, code int) {
	errParams := RespParams{Error: errMsg}
	data, err := json.Marshal(errParams)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

func respJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}
