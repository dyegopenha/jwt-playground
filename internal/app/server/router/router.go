package router

import (
	"net/http"

	"github.com/dyegopenha/jwt-playground/internal/app/server/handler"
	"github.com/dyegopenha/jwt-playground/internal/app/server/middleware"
)

type Router struct {
	*http.ServeMux

	m  *middleware.Middleware
	ah *handler.AuthHandler
	uh *handler.UserHandler
}

// NewMux assembles the HTTP routes and returns a ready-to-use ServeMux.
func NewRouter(
	m *middleware.Middleware,
	ah *handler.AuthHandler,
	uh *handler.UserHandler,
) *Router {
	mux := http.NewServeMux()

	return &Router{
		ServeMux: mux,
		m:        m,
		ah:       ah,
		uh:       uh,
	}
}

func (r *Router) Register() {
	// Public endpoints
	r.Handle("/sign-in", http.HandlerFunc(r.ah.SignIn))
	r.Handle("/refresh", http.HandlerFunc(r.ah.Refresh))

	// Protected endpoints
	r.Handle(
		"/",
		r.m.JWTMiddleware(http.HandlerFunc(r.uh.Profile)),
	)
}
