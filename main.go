package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// -----------------------------------------------------------------------------
// 1. Domain-specific data held in the token
// -----------------------------------------------------------------------------
type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

// -----------------------------------------------------------------------------
// 2. Signing key (HS256 demo – use env / KMS in real code)
// -----------------------------------------------------------------------------
var hmacKey = []byte("change-me-in-prod")

// -----------------------------------------------------------------------------
// 2b. Refresh-token store and helpers (demo-only, in-memory)
// -----------------------------------------------------------------------------

type RefreshSession struct {
	UserID    string
	Role      string
	ExpiresAt time.Time
}

var (
	refreshTokens = make(map[string]RefreshSession)
	refreshMu     sync.Mutex
)

// generateRandomBase64 returns a URL-safe base64 string that encodes n random
// bytes (no padding).
func generateRandomBase64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).
			EncodeToString(b),
		nil
}

// GenerateRefreshToken issues a refresh token and stores its session data.
func GenerateRefreshToken(
	userID, role string,
	ttl time.Duration,
) (string, error) {
	tok, err := generateRandomBase64(32)
	if err != nil {
		return "", err
	}
	refreshMu.Lock()
	refreshTokens[tok] = RefreshSession{
		UserID:    userID,
		Role:      role,
		ExpiresAt: time.Now().Add(ttl),
	}
	refreshMu.Unlock()
	return tok, nil
}

// IssueTokenPair returns an access token plus a refresh token.
func IssueTokenPair(
	userID, role string,
	accessTTL, refreshTTL time.Duration,
) (string, string, error) {
	accessTok, err := SignAccessToken(userID, role, accessTTL)
	if err != nil {
		return "", "", err
	}
	refreshTok, err := GenerateRefreshToken(userID, role, refreshTTL)
	if err != nil {
		return "", "", err
	}
	return accessTok, refreshTok, nil
}

// -----------------------------------------------------------------------------
// 3. Create & sign an access token
// -----------------------------------------------------------------------------

func SignAccessToken(
	userID, role string,
	ttl time.Duration,
) (string, error) {
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: userID,
			ExpiresAt: jwt.
				NewNumericDate(time.Now().Add(ttl)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.
		NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(hmacKey)
}

// -----------------------------------------------------------------------------
// 4. Parse & verify a token string → claims
// -----------------------------------------------------------------------------

func ParseAndVerify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			return hmacKey, nil
		},
		jwt.WithValidMethods(
			[]string{jwt.SigningMethodHS256.Alg()},
		),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

// -----------------------------------------------------------------------------
// 5. Middleware: validate "Bearer <token>" and put claims in context
// -----------------------------------------------------------------------------
type ctxKey string

const claimsKey ctxKey = "jwtClaims"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing bearer token",
				http.StatusUnauthorized)
			return
		}
		raw := strings.TrimPrefix(auth, "Bearer ")
		claims, err := ParseAndVerify(raw)
		if err != nil {
			http.Error(w, "invalid or expired token",
				http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(),
			claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// Helper to pull claims inside handlers
func CurrentUser(r *http.Request) *Claims {
	if v, ok := r.Context().Value(claimsKey).(*Claims); ok {
		return v
	}
	return nil
}

// -----------------------------------------------------------------------------
// 6. Refresh endpoint
// -----------------------------------------------------------------------------
const refreshCookieName = "refresh_token"

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		http.Error(w, "missing refresh token cookie", http.StatusUnauthorized)
		return
	}
	rawRefreshTok := cookie.Value

	refreshMu.Lock()
	session, ok := refreshTokens[rawRefreshTok]
	refreshMu.Unlock()
	if !ok || time.Now().After(session.ExpiresAt) {
		http.Error(
			w,
			"invalid or expired refresh token",
			http.StatusUnauthorized,
		)
		return
	}

	accessTok, newRefreshTok, err := IssueTokenPair(
		session.UserID,
		session.Role,
		15*time.Minute, // access-token TTL
		24*time.Hour,   // refresh-token TTL
	)
	if err != nil {
		http.Error(w, "token generation error", http.StatusInternalServerError)
		return
	}

	refreshMu.Lock()
	delete(refreshTokens, rawRefreshTok)
	refreshMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    newRefreshTok,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true, // set to false if you are not using TLS in dev
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessTok,
	}); err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
	}
}

func main() {
	accessTok, refreshTok, _ := IssueTokenPair(
		"user-123",
		"admin",
		15*time.Minute,
		24*time.Hour,
	)
	log.Println("example access token:", accessTok)
	log.Println("example refresh token:", refreshTok)

	mux := http.NewServeMux()
	mux.Handle("/refresh", http.HandlerFunc(RefreshHandler))
	mux.Handle(
		"/",
		AuthMiddleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c := CurrentUser(r)
				_, err := w.Write([]byte("Hello, " + c.Issuer + "!"))
				if err != nil {
					log.Println("error writing response:", err)
				}
			}),
		),
	)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
