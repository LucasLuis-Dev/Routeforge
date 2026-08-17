package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type RideStatus string

const (
	StatusRequested  RideStatus = "requested"
	StatusAccepted   RideStatus = "accepted"
	StatusInProgress RideStatus = "in_progress"
	StatusCompleted  RideStatus = "completed"
	StatusCanceled   RideStatus = "canceled"
)

var (
	ErrRideNotFound            = errors.New("corrida não encontrada")
	ErrInvalidStatusTransition = errors.New("transição de status de corrida inválida")
	ErrDriverAlreadyAssigned   = errors.New("corrida já possui motorista atribuído")
)

type Ride struct {
	ID                   uuid.UUID  `json:"id"`
	PassengerID          uuid.UUID  `json:"passenger_id"`
	DriverID             *uuid.UUID `json:"driver_id,omitempty"`
	OriginLatitude       float64    `json:"origin_latitude"`
	OriginLongitude      float64    `json:"origin_longitude"`
	DestinationLatitude  float64    `json:"destination_latitude"`
	DestinationLongitude float64    `json:"destination_longitude"`
	DistanceKM           float64    `json:"distance_km"`
	Status               RideStatus `json:"status"`
	EstimatedPrice       float64    `json:"estimated_price"`
	FinalPrice           *float64   `json:"final_price,omitempty"`
	ETAMinutes           int        `json:"eta_minutes"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type PriceHistory struct {
	ID                   uuid.UUID `json:"id"`
	RideID               uuid.UUID `json:"ride_id"`
	BaseFare             float64   `json:"base_fare"`
	DistanceFare         float64   `json:"distance_fare"`
	SurgeMultiplier      float64   `json:"surge_multiplier"`
	FinalCalculatedPrice float64   `json:"final_calculated_price"`
	IsFallback           bool      `json:"is_fallback"`
	CalculatedAt         time.Time `json:"calculated_at"`
}

type RideRepository interface {
	Create(ctx context.Context, ride *Ride) error
	CreateRideWithOutbox(ctx context.Context, ride *Ride, history *PriceHistory, outbox *OutboxEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*Ride, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status RideStatus, driverID *uuid.UUID, finalPrice *float64) error
	UpdateStatusWithOutbox(ctx context.Context, id uuid.UUID, status RideStatus, driverID *uuid.UUID, finalPrice *float64, outbox *OutboxEvent) error
	SavePriceHistory(ctx context.Context, history *PriceHistory) error
	GetPriceHistoryByRideID(ctx context.Context, rideID uuid.UUID) ([]*PriceHistory, error)
}
