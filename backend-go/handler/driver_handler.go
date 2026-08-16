package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	redisRepo "github.com/LucasLuis-Dev/Routeforge/backend-go/repository/redis"
)

type UpdateLocationRequest struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type DriverHandler struct {
	geoRepo redisRepo.GeoRepository
}

func NewDriverHandler(geoRepo redisRepo.GeoRepository) *DriverHandler {
	return &DriverHandler{geoRepo: geoRepo}
}

func (h *DriverHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var req UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "corpo da requisição JSON inválido")
		return
	}

	if req.DriverID == "" || req.Latitude == 0 || req.Longitude == 0 {
		respondError(w, http.StatusUnprocessableEntity, "driver_id, latitude e longitude são obrigatórios")
		return
	}

	if err := h.geoRepo.UpdateDriverLocation(r.Context(), req.DriverID, req.Latitude, req.Longitude); err != nil {
		respondError(w, http.StatusInternalServerError, "erro ao atualizar posição GPS do motorista no Redis")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message":   "localização GPS do motorista atualizada com sucesso no Redis GEO",
		"driver_id": req.DriverID,
	})
}

func (h *DriverHandler) FindNearby(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("latitude")
	lngStr := r.URL.Query().Get("longitude")
	radiusStr := r.URL.Query().Get("radius")
	countStr := r.URL.Query().Get("count")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "parâmetro 'latitude' numérico é obrigatório")
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "parâmetro 'longitude' numérico é obrigatório")
		return
	}

	radiusKm := 5.0
	if radiusStr != "" {
		if rKm, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			radiusKm = rKm
		}
	}

	count := 5
	if countStr != "" {
		if c, err := strconv.Atoi(countStr); err == nil {
			count = c
		}
	}

	drivers, err := h.geoRepo.FindNearbyDrivers(r.Context(), lat, lng, radiusKm, count)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "erro ao consultar motoristas próximos no Redis")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total":   len(drivers),
		"drivers": drivers,
	})
}
