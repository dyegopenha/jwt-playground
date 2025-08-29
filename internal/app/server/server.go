package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dyegopenha/jwt-playground/internal/app/server/router"
	"github.com/dyegopenha/jwt-playground/internal/config/env"
)

type Server struct {
	e *env.Env
	r *router.Router
}

func newServer(
	e *env.Env,
	r *router.Router,
) *Server {
	return &Server{
		e: e,
		r: r,
	}
}

func (s *Server) Run(ctx context.Context) error {
	log.Printf("starting server on port %s", s.e.Port)

	srv := &http.Server{
		Addr:    ":" + s.e.Port,
		Handler: s.r.ServeMux,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			errCh <- fmt.Errorf("listen: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return ctx.Err()

	case err := <-errCh:
		return err
	}
}
