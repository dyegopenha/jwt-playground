package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dyegopenha/jwt-playground/internal/config/env"
	"github.com/dyegopenha/jwt-playground/internal/domain/usecase"
)

type AuthHandler struct {
	e   *env.Env
	sic *usecase.SignInUseCase
	ruc *usecase.RefreshUseCase
}

func NewAuthHandler(
	e *env.Env,
	sic *usecase.SignInUseCase,
	ruc *usecase.RefreshUseCase,
) *AuthHandler {
	return &AuthHandler{
		e:   e,
		sic: sic,
		ruc: ruc,
	}
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	accessTok, refreshTok, err := h.sic.Execute(
		r.Context(),
		creds.Email,
		creds.Password,
	)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   refreshCookieName,
		Value:  refreshTok,
		Path:   "/",
		MaxAge: int(h.e.RefreshTokenTTL.Seconds()),
	})

	if err := json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessTok,
	}); err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		http.Error(w, "missing refresh token cookie", http.StatusUnauthorized)
		return
	}

	refreshTok := cookie.Value
	accessTok, newRefreshTok, err := h.ruc.Execute(r.Context(), refreshTok)
	if err != nil {
		http.Error(w, "failed to refresh token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   refreshCookieName,
		Value:  newRefreshTok,
		Path:   "/",
		MaxAge: int(h.e.RefreshTokenTTL.Seconds()),
	})

	if err := json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessTok,
	}); err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		return
	}
}
