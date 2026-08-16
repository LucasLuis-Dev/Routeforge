package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
)

type UserHandler struct {
	userRepo domain.UserRepository
}

func NewUserHandler(userRepo domain.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

type CreateUserRequest struct {
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	UserType domain.UserType `json:"user_type"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "corpo da requisição JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" {
		http.Error(w, `{"error": "nome e email são obrigatórios"}`, http.StatusUnprocessableEntity)
		return
	}

	if req.UserType != domain.UserTypePassenger && req.UserType != domain.UserTypeDriver {
		http.Error(w, `{"error": "user_type deve ser 'passenger' ou 'driver'"}`, http.StatusUnprocessableEntity)
		return
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		UserType: req.UserType,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			http.Error(w, `{"error": "email já cadastrado"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error": "erro interno ao cadastrar usuário"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}
