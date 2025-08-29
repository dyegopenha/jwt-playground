package middleware

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const (
	ClaimsKey ctxKey = "jwt_claims"
)

// AuthMiddleware verifies the Authorization header ("Bearer <token>").
// On success it attaches the jwtutil.Claims to the request context so that
// downstream handlers can retrieve the current user via CurrentUser.
func (m *Middleware) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		raw := strings.TrimPrefix(auth, "Bearer ")

		claims, err := m.j.ParseAndVerify(raw)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
