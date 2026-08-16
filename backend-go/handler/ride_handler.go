package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RideHandler struct {
	rideService service.RideService
}

func NewRideHandler(rideService service.RideService) *RideHandler {
	return &RideHandler{rideService: rideService}
}

func (h *RideHandler) EstimateRide(w http.ResponseWriter, r *http.Request) {
	var req service.EstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "corpo da requisição JSON inválido"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.rideService.CreateEstimate(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RideHandler) CreateRide(w http.ResponseWriter, r *http.Request) {
	var req service.CreateRideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "corpo da requisição JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if req.PassengerID == uuid.Nil {
		http.Error(w, `{"error": "passenger_id é obrigatório"}`, http.StatusUnprocessableEntity)
		return
	}

	ride, err := h.rideService.CreateRide(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, `{"error": "passageiro não encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ride)
}

type AcceptRideRequest struct {
	DriverID uuid.UUID `json:"driver_id"`
}

func (h *RideHandler) AcceptRide(w http.ResponseWriter, r *http.Request) {
	rideIDParam := chi.URLParam(r, "id")
	rideID, err := uuid.Parse(rideIDParam)
	if err != nil {
		http.Error(w, `{"error": "ID de corrida inválido"}`, http.StatusBadRequest)
		return
	}

	var req AcceptRideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DriverID == uuid.Nil {
		http.Error(w, `{"error": "driver_id válido é obrigatório"}`, http.StatusBadRequest)
		return
	}

	ride, err := h.rideService.AcceptRide(r.Context(), rideID, req.DriverID)
	if err != nil {
		if errors.Is(err, domain.ErrRideNotFound) || errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ride)
}

func (h *RideHandler) CompleteRide(w http.ResponseWriter, r *http.Request) {
	rideIDParam := chi.URLParam(r, "id")
	rideID, err := uuid.Parse(rideIDParam)
	if err != nil {
		http.Error(w, `{"error": "ID de corrida inválido"}`, http.StatusBadRequest)
		return
	}

	ride, err := h.rideService.CompleteRide(r.Context(), rideID)
	if err != nil {
		if errors.Is(err, domain.ErrRideNotFound) {
			http.Error(w, `{"error": "corrida não encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ride)
}

func (h *RideHandler) GetRideDetails(w http.ResponseWriter, r *http.Request) {
	rideIDParam := chi.URLParam(r, "id")
	rideID, err := uuid.Parse(rideIDParam)
	if err != nil {
		http.Error(w, `{"error": "ID de corrida inválido"}`, http.StatusBadRequest)
		return
	}

	details, err := h.rideService.GetRideDetails(r.Context(), rideID)
	if err != nil {
		if errors.Is(err, domain.ErrRideNotFound) {
			http.Error(w, `{"error": "corrida não encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "erro ao buscar detalhes da corrida"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(details)
}
