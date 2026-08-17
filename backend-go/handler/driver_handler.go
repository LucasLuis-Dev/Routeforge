package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/pkg/messaging"
	redisRepo "github.com/LucasLuis-Dev/Routeforge/backend-go/repository/redis"
	ws "github.com/LucasLuis-Dev/Routeforge/backend-go/websocket"
)

type UpdateLocationRequest struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type DriverHandler struct {
	geoRepo        redisRepo.GeoRepository
	eventPublisher messaging.EventPublisher
	wsHub          *ws.Hub
}

func NewDriverHandler(geoRepo redisRepo.GeoRepository, eventPublisher messaging.EventPublisher, wsHub *ws.Hub) *DriverHandler {
	return &DriverHandler{
		geoRepo:        geoRepo,
		eventPublisher: eventPublisher,
		wsHub:          wsHub,
	}
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

	// 1. Atualiza no Redis GEO
	if err := h.geoRepo.UpdateDriverLocation(r.Context(), req.DriverID, req.Latitude, req.Longitude); err != nil {
		respondError(w, http.StatusInternalServerError, "erro ao atualizar posição GPS do motorista no Redis")
		return
	}

	locationPayload := map[string]interface{}{
		"driver_id": req.DriverID,
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"timestamp": time.Now(),
	}

	// 2. Dispara evento assíncrono no RabbitMQ: driver.location_updated
	if h.eventPublisher != nil {
		_ = h.eventPublisher.PublishEvent(r.Context(), "driver.location_updated", locationPayload)
	}

	// 3. Transmite em tempo real via WebSocket Streaming Gateway para passageiros em escuta
	if h.wsHub != nil {
		h.wsHub.BroadcastEvent("driver_location_update", locationPayload)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "localização GPS do motorista atualizada com sucesso no Redis GEO, RabbitMQ & WebSockets",
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
