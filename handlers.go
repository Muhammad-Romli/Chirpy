package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Muhammad-Romli/Chirpy/internal/auth"
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

type userRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type createChirpRequest struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Fail writing response: %v", err)
	}
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
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

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		respError(w, fmt.Sprintf("Fail writing response: %v", err), 500)
	}
	err = cfg.DbQueries.ResetUsers(r.Context())
	if err != nil {
		respError(w, fmt.Sprintf("fail resetting users in database: %v", err), 500)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileServerHits.Add(1)
		handler.ServeHTTP(w, r)
	})
}

func decodeReq[T any](w http.ResponseWriter, r *http.Request, params *T) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(params); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
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

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	createUserParams := userRequest{}
	err := decodeReq(w, r, &createUserParams)
	if err != nil {
		respError(w, fmt.Sprintln("fail decoding request:", err), 500)
		return
	}
	hashed_pw, err := auth.HashPassword(createUserParams.Password)
	if err != nil {
		respError(w, fmt.Sprintln("fail hashing password", err), 500)
		return
	}
	userParams := database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          createUserParams.Email,
		HashedPassword: hashed_pw,
	}
	user, err := cfg.DbQueries.CreateUser(r.Context(), userParams)
	if err != nil {
		respError(w, fmt.Sprintf("Fail creating user in database: %v", err), 500)
		return
	}
	respJSON(w, 201, user)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	userReq := userRequest{}
	err := decodeReq(w, r, &userReq)
	if err != nil {
		respError(w, fmt.Sprintln("fail decoding request:", err), 500)
		return
	}

	user, err := cfg.DbQueries.GetUserByEmail(r.Context(), userReq.Email)
	if err != nil {
		respError(w, "incorrect email or password", 401)
		return
	}

	match, err := auth.CheckPasswordHash(userReq.Password, user.HashedPassword)
	if err != nil || !match {
		respError(w, "incorrect email or password", 401)
		return
	}

	expiresIn := time.Hour
	if userReq.ExpiresInSeconds > 0 {
		requested := time.Duration(userReq.ExpiresInSeconds) * time.Second
		if requested < time.Hour {
			expiresIn = requested
		}
	}

	token, err := auth.MakeJWT(user.ID, cfg.SecretJWT, expiresIn)
	if err != nil {
		respError(w, "couldn't create token", 500)
		return
	}

	respJSON(w, 200, userResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})
}

func (cfg *apiConfig) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := cfg.DbQueries.GetUsers(r.Context())
	if err != nil {
		respError(w, "Fail getting users from db", 500)
		return
	}
	respJSON(w, 200, users)
}

func (cfg *apiConfig) getUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("userID")
	id, err := uuid.Parse(stringID)
	if err != nil {
		respError(w, "Cannot parse given ID", 404)
		return
	}
	user, err := cfg.DbQueries.GetUser(r.Context(), id)
	if err != nil {
		respError(w, "Fail getting user from db", 500)
		return
	}
	respJSON(w, 200, user)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	createChirpParams := createChirpRequest{}
	err := decodeReq(w, r, &createChirpParams)
	if err != nil {
		respError(w, fmt.Sprintln("fail decoding request:", err), 500)
		return
	}
	validateChirp(w, createChirpParams)
	userID, err := uuid.Parse(createChirpParams.UserID)

	if err != nil {
		respError(w, "Cannot parse user ID", 400)
		return
	}
	chirpParams := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      createChirpParams.Body,
		UserID:    userID,
	}
	chirp, err := cfg.DbQueries.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		respError(w, "Fail creating chirp", 500)
		return
	}
	respJSON(w, 201, chirp)
}

func validateChirp(w http.ResponseWriter, params createChirpRequest) {
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

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respError(w, fmt.Sprintf("fail getting token from bearer: %v", err), http.StatusBadRequest)
		return
	}
	_, err = auth.ValidateJWT(token, cfg.SecretJWT)
	if err != nil {
		respError(w, "JWT is invalid", 401)
		return
	}
	chirps, err := cfg.DbQueries.GetChirps(r.Context())
	if err != nil {
		respError(w, "Fail getting chirps from db", 500)
		return
	}
	respJSON(w, 200, chirps)
}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("chirpID")
	id, err := uuid.Parse(stringID)
	if err != nil {
		respError(w, "Cannot parse given ID", 404)
		return
	}
	chirp, err := cfg.DbQueries.GetChirp(r.Context(), id)
	if err != nil {
		respError(w, "Fail getting chirp from db", 500)
		return
	}
	respJSON(w, 200, chirp)
}
