package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Muhammad-Romli/Chirpy/internal/database"
	"github.com/google/uuid"
)

type ReqParams struct {
	Body string `json:"body"`
}

type ErrParams struct {
	Error string `json:"error"`
}

type RespParams struct {
	CleanedBody string `json:"cleaned_body"`
}

type createUserRequest struct {
	Email string `json:"email"`
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
</html>`, cfg.FileServerHits.Load())
	_, err := w.Write([]byte(countStr))
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Fail writing response: %v", err)
	}
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	count := cfg.FileServerHits.Load()
	cfg.FileServerHits.Add(-count)
	_, err := w.Write([]byte("FileServerHits count reseted."))
	respError(w, fmt.Sprintf("Fail writing response: %v", err), 500)
	err = cfg.DbQueries.ResetUsers(req.Context())
	respError(w, fmt.Sprintf("fail resetting users in database: %v", err), 500)
}

func (cfg *apiConfig) middlewareMetricsInc(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.FileServerHits.Add(1)
		handler.ServeHTTP(w, req)
	})
}

func validateChirpHandler(w http.ResponseWriter, req *http.Request) {
	params := ReqParams{}
	decodeReq(w, req, params)
	if len(params.Body) > 140 {
		errMsg := "Chirp is too long"
		respError(w, errMsg, 400)
		return
	}
	words := strings.Split(params.Body, " ")
	finishWords := []string{}
	for _, w := range words {
		w = strings.ToLower(w)
		if w == "kerfuffle" || w == "sharbert" || w == "fornax" {
			w = "****"
		}
		finishWords = append(finishWords, w)
	}
	finishChirps := strings.Join(finishWords, " ")
	resp := RespParams{CleanedBody: finishChirps}
	respJSON(w, 200, resp)
}

func decodeReq(w http.ResponseWriter, req *http.Request, params interface{}) {
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
}

func respError(w http.ResponseWriter, errMsg string, code int) {
	log.Print(errMsg)
	params := ErrParams{Error: errMsg}
	data, err := json.Marshal(params)
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
	log.Printf("Successfully responding with JSON")
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, req *http.Request) {
	createUserParams := createUserRequest{}
	decodeReq(w, req, createUserParams)
	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     createUserParams.Email,
	}
	user, err := cfg.DbQueries.CreateUser(req.Context(), userParams)
	if err != nil {
		respError(w, fmt.Sprintf("fail creating user in database: %v", err), 500)
	}
	respJSON(w, 201, user)
}
