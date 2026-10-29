package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashPw, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("fail hashing password: %w", err)
	}
	return hashPw, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("fail comparing hash and pw: %w", err)
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	currentTime := time.Now()
	regClaim := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, regClaim)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("fail signing jwt: %w", err)
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	getSecret := func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	}

	parsedToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, getSecret)
	if err != nil {
		return uuid.Nil, fmt.Errorf("fail parsing token: %w", err)
	}
	if !parsedToken.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid token claims")
	}

	subject := claims.Subject
	parsedUUID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("fail parsing subject uuid: %w", err)
	}
	return parsedUUID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	value := headers.Get("Authorization")
	if value == "" {
		return "", errors.New("missing Authorization header")
	}
	token, ok := strings.CutPrefix(value, "Bearer ")
	if !ok {
		return "", errors.New("malformed Authorization header")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("token is empty")
	}
	return token, nil
}

func MakeRefreshToken() (string, error) {
	random := make([]byte, 32)
	_, err := rand.Read(random)
	if err != nil {
		return "", fmt.Errorf("Fail generating random value: %w", err)
	}
	return hex.EncodeToString(random), nil
}
