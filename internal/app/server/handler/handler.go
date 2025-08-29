package handler

import (
	"net/http"

	"github.com/dyegopenha/jwt-playground/internal/app/server/middleware"
	"github.com/dyegopenha/jwt-playground/internal/pkg/jwtutil"
)

const refreshCookieName string = "refresh_token"

func CurrentUser(r *http.Request) *jwtutil.Claims {
	if v, ok := r.Context().Value(middleware.ClaimsKey).(*jwtutil.Claims); ok {
		return v
	}
	return nil
}
