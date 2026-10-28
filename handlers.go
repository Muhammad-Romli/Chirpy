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
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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
		respError(w, "Fail writing response", 500, err)
	}
	err = cfg.DbQueries.ResetUsers(r.Context())
	if err != nil {
		respError(w, "fail resetting users in database", 500, err)
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

func respError(w http.ResponseWriter, msg string, code int, err error) {
	if err != nil {
		log.Printf("%s: %v", msg, err) // full detail, server only
	} else {
		log.Print(msg)
	}
	data, mErr := json.Marshal(ErrParams{Error: msg}) // client only sees msg
	if mErr != nil {
		log.Printf("Error marshalling JSON: %s", mErr)
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}
	respJSON(w, code, data)
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
		respError(w, "fail decoding request:", 500, err)
		return
	}
	hashed_pw, err := auth.HashPassword(createUserParams.Password)
	if err != nil {
		respError(w, "fail hashing password", 500, err)
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
		respError(w, "Fail creating user in database", 500, err)
		return
	}
	respJSON(w, 201, user)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	userReq := userRequest{}
	err := decodeReq(w, r, &userReq)
	if err != nil {
		respError(w, "fail decoding request:", 500, err)
		return
	}

	user, err := cfg.DbQueries.GetUserByEmail(r.Context(), userReq.Email)
	if err != nil {
		respError(w, "incorrect email or password", 401, err)
		return
	}

	match, err := auth.CheckPasswordHash(userReq.Password, user.HashedPassword)
	if err != nil || !match {
		respError(w, "incorrect email or password", 401, err)
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
		respError(w, "Fail creating JWT token", 500, err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respError(w, "Fail creating refresh token", 500, err)
		return
	}

	params := database.CreateRefreshTokenParams{
		RefreshToken: refreshToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		UserID:       user.ID,
		ExpiresAt:    time.Now().Add(60 * 24 * time.Hour),
	}

	_, err = cfg.DbQueries.CreateRefreshToken(r.Context(), params)
	if err != nil {
		respError(w, "fail storing refresh token to database", 500, err)
		return
	}

	respJSON(w, 200, userResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := cfg.DbQueries.GetUsers(r.Context())
	if err != nil {
		respError(w, "Fail getting users from db", 500, err)
		return
	}
	respJSON(w, 200, users)
}

func (cfg *apiConfig) getUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("userID")
	id, err := uuid.Parse(stringID)
	if err != nil {
		respError(w, "Cannot parse given ID", 404, err)
		return
	}
	user, err := cfg.DbQueries.GetUser(r.Context(), id)
	if err != nil {
		respError(w, "Fail getting user from db", 500, err)
		return
	}
	respJSON(w, 200, user)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	createChirpParams := createChirpRequest{}
	err := decodeReq(w, r, &createChirpParams)
	if err != nil {
		respError(w, "fail decoding request", 500, err)
		return
	}
	validateChirp(w, createChirpParams)
	userID, err := uuid.Parse(createChirpParams.UserID)

	if err != nil {
		respError(w, "Cannot parse user ID", 400, err)
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
		respError(w, "Fail creating chirp", 500, err)
		return
	}
	respJSON(w, 201, chirp)
}

func validateChirp(w http.ResponseWriter, params createChirpRequest) {
	if len(params.Body) > 140 {
		respError(w, "Chirp is too long", 400, fmt.Errorf("message is too long"))
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
		respError(w, "fail getting token from bearer", http.StatusBadRequest, err)
		return
	}
	_, err = auth.ValidateJWT(token, cfg.SecretJWT)
	if err != nil {
		respError(w, "JWT is invalid", 401, err)
		return
	}
	chirps, err := cfg.DbQueries.GetChirps(r.Context())
	if err != nil {
		respError(w, "Fail getting chirps from db", 500, err)
		return
	}
	respJSON(w, 200, chirps)
}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("chirpID")
	id, err := uuid.Parse(stringID)
	if err != nil {
		respError(w, "Cannot parse given ID", 404, err)
		return
	}
	chirp, err := cfg.DbQueries.GetChirp(r.Context(), id)
	if err != nil {
		respError(w, "Fail getting chirp from db", 500, err)
		return
	}
	respJSON(w, 200, chirp)
}
