package handler

import (
	"encoding/json"
	"net/http"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/pkg/auth"
)

type LoginRequest struct {
	Email string `json:"email"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

type AuthHandler struct {
	userRepo domain.UserRepository
}

func NewAuthHandler(userRepo domain.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "formato de JSON inválido")
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "o email é obrigatório")
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusNotFound, "usuário não encontrado com o email informado")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.UserType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "erro ao gerar token JWT")
		return
	}

	respondJSON(w, http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
