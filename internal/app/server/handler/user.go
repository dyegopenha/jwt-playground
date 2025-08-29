package handler

import (
	"encoding/json"
	"net/http"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (uh *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"id":   user.Issuer,
		"role": user.Role,
	}); err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		return
	}
}
