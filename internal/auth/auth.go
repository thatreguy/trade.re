package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

// Auth handles authentication operations
type Auth struct {
	jwtSecret    []byte
	tokenExpiry  time.Duration
	apiKeyLength int
}

// Claims represents JWT claims
type Claims struct {
	TraderID uuid.UUID `json:"trader_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// New creates a new Auth instance
func New(jwtSecret string, tokenExpiryHours int, apiKeyLength int) *Auth {
	return &Auth{
		jwtSecret:    []byte(jwtSecret),
		tokenExpiry:  time.Duration(tokenExpiryHours) * time.Hour,
		apiKeyLength: apiKeyLength,
	}
}

// HashPassword hashes a password using bcrypt
func (a *Auth) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword checks if a password matches a hash
func (a *Auth) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken creates a JWT token for a trader
func (a *Auth) GenerateToken(traderID uuid.UUID, username string) (string, error) {
	claims := &Claims{
		TraderID: traderID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "trade.re",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return tokenString, nil
}

// ValidateToken verifies and parses a JWT token
func (a *Auth) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateAPIKey creates a new API key
func (a *Auth) GenerateAPIKey() string {
	bytes := make([]byte, a.apiKeyLength)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// HashAPIKey creates a hash of an API key for storage
func (a *Auth) HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// ExtractToken extracts the Bearer token from an Authorization header
func ExtractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// ExtractAPIKey extracts the API key from X-API-Key header
func ExtractAPIKey(r *http.Request) string {
	return r.Header.Get("X-API-Key")
}
