package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/pkg/geo"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/pkg/messaging"
	redisRepo "github.com/LucasLuis-Dev/Routeforge/backend-go/repository/redis"
	"github.com/google/uuid"
)

const (
	DefaultBaseFare  = 2.50
	DefaultRatePerKM = 1.80
	DefaultCitySpeed = 30.0 // km/h
)

type EstimateRequest struct {
	OriginLatitude       float64 `json:"origin_latitude"`
	OriginLongitude      float64 `json:"origin_longitude"`
	DestinationLatitude  float64 `json:"destination_latitude"`
	DestinationLongitude float64 `json:"destination_longitude"`
}

type EstimateResponse struct {
	DistanceKM      float64 `json:"distance_km"`
	ETAMinutes      int     `json:"eta_minutes"`
	BaseFare        float64 `json:"base_fare"`
	DistanceFare    float64 `json:"distance_fare"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
	EstimatedPrice  float64 `json:"estimated_price"`
	IsFallback      bool    `json:"is_fallback"`
	Cached          bool    `json:"cached,omitempty"`
}

type CreateRideRequest struct {
	PassengerID          uuid.UUID `json:"passenger_id"`
	OriginLatitude       float64   `json:"origin_latitude"`
	OriginLongitude      float64   `json:"origin_longitude"`
	DestinationLatitude  float64   `json:"destination_latitude"`
	DestinationLongitude float64   `json:"destination_longitude"`
}

type RideDetailsResponse struct {
	Ride         *domain.Ride           `json:"ride"`
	PriceHistory []*domain.PriceHistory `json:"price_history"`
}

type RideService interface {
	CreateEstimate(ctx context.Context, req EstimateRequest) (*EstimateResponse, error)
	CreateRide(ctx context.Context, req CreateRideRequest) (*domain.Ride, error)
	AcceptRide(ctx context.Context, rideID uuid.UUID, driverID uuid.UUID) (*domain.Ride, error)
	CompleteRide(ctx context.Context, rideID uuid.UUID) (*domain.Ride, error)
	GetRideDetails(ctx context.Context, rideID uuid.UUID) (*RideDetailsResponse, error)
}

type rideService struct {
	userRepo       domain.UserRepository
	rideRepo       domain.RideRepository
	mlClient       domain.PredictionClient
	estimateCache  redisRepo.EstimateCache
	eventPublisher messaging.EventPublisher
}

func NewRideService(
	userRepo domain.UserRepository,
	rideRepo domain.RideRepository,
	mlClient domain.PredictionClient,
	estimateCache redisRepo.EstimateCache,
	eventPublisher messaging.EventPublisher,
) RideService {
	return &rideService{
		userRepo:       userRepo,
		rideRepo:       rideRepo,
		mlClient:       mlClient,
		estimateCache:  estimateCache,
		eventPublisher: eventPublisher,
	}
}

func (s *rideService) CreateEstimate(ctx context.Context, req EstimateRequest) (*EstimateResponse, error) {
	if req.OriginLatitude == 0 && req.OriginLongitude == 0 {
		return nil, errors.New("coordenadas de origem inválidas")
	}
	if req.DestinationLatitude == 0 && req.DestinationLongitude == 0 {
		return nil, errors.New("coordenadas de destino inválidas")
	}

	// 1. Tentar obter a estimativa do Redis Cache (Cache HIT)
	if s.estimateCache != nil {
		if cached, err := s.estimateCache.GetEstimate(ctx, req.OriginLatitude, req.OriginLongitude, req.DestinationLatitude, req.DestinationLongitude); err == nil && cached != nil {
			return &EstimateResponse{
				DistanceKM:      cached.DistanceKM,
				ETAMinutes:      cached.ETAMinutes,
				BaseFare:        cached.BaseFare,
				DistanceFare:    cached.DistanceFare,
				SurgeMultiplier: cached.SurgeMultiplier,
				EstimatedPrice:  cached.EstimatedPrice,
				IsFallback:      cached.IsFallback,
				Cached:          true,
			}, nil
		}
	}

	// 2. Cache MISS: Calcular distância e solicitar ao modelo de ML
	distanceKM := geo.CalculateHaversineDistance(
		req.OriginLatitude, req.OriginLongitude,
		req.DestinationLatitude, req.DestinationLongitude,
	)

	now := time.Now()
	predReq := &domain.PredictionRequest{
		DistanceKM: distanceKM,
		HourOfDay:  now.Hour(),
		DayOfWeek:  int(now.Weekday()),
	}

	var resp *EstimateResponse

	predResp, err := s.mlClient.Predict(ctx, predReq)
	if err != nil {
		// FALLBACK PATTERN: Acionado por erro HTTP/Timeout ou por Circuit Breaker (gobreaker.ErrOpenState)
		fmt.Printf("Fallback acionado no cálculo de corrida (Motivo: %v)\n", err)

		fallbackDistanceFare := math.Round(distanceKM*DefaultRatePerKM*100) / 100
		fallbackPrice := math.Round((DefaultBaseFare+fallbackDistanceFare)*100) / 100
		fallbackETA := int(math.Max(2, math.Round(distanceKM/DefaultCitySpeed*60.0)))

		resp = &EstimateResponse{
			DistanceKM:      distanceKM,
			ETAMinutes:      fallbackETA,
			BaseFare:        DefaultBaseFare,
			DistanceFare:    fallbackDistanceFare,
			SurgeMultiplier: 1.00,
			EstimatedPrice:  fallbackPrice,
			IsFallback:      true,
			Cached:          false,
		}
	} else {
		resp = &EstimateResponse{
			DistanceKM:      distanceKM,
			ETAMinutes:      predResp.ETAMinutes,
			BaseFare:        predResp.BaseFare,
			DistanceFare:    predResp.DistanceFare,
			SurgeMultiplier: predResp.SurgeMultiplier,
			EstimatedPrice:  predResp.EstimatedPrice,
			IsFallback:      false,
			Cached:          false,
		}
	}

	// 3. Salvar no Redis Cache por 3 minutos (TTL 180s)
	if s.estimateCache != nil && !resp.IsFallback {
		_ = s.estimateCache.SetEstimate(ctx, req.OriginLatitude, req.OriginLongitude, req.DestinationLatitude, req.DestinationLongitude, &redisRepo.CachedEstimate{
			DistanceKM:      resp.DistanceKM,
			ETAMinutes:      resp.ETAMinutes,
			BaseFare:        resp.BaseFare,
			DistanceFare:    resp.DistanceFare,
			SurgeMultiplier: resp.SurgeMultiplier,
			EstimatedPrice:  resp.EstimatedPrice,
			IsFallback:      resp.IsFallback,
		}, 3*time.Minute)
	}

	return resp, nil
}

func (s *rideService) CreateRide(ctx context.Context, req CreateRideRequest) (*domain.Ride, error) {
	passenger, err := s.userRepo.GetByID(ctx, req.PassengerID)
	if err != nil {
		return nil, fmt.Errorf("passageiro não encontrado: %w", err)
	}
	if passenger.UserType != domain.UserTypePassenger {
		return nil, errors.New("apenas usuários do tipo 'passenger' podem solicitar corridas")
	}

	estimate, err := s.CreateEstimate(ctx, EstimateRequest{
		OriginLatitude:       req.OriginLatitude,
		OriginLongitude:      req.OriginLongitude,
		DestinationLatitude:  req.DestinationLatitude,
		DestinationLongitude: req.DestinationLongitude,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular estimativa: %w", err)
	}

	ride := &domain.Ride{
		ID:                   uuid.New(),
		PassengerID:          req.PassengerID,
		OriginLatitude:       req.OriginLatitude,
		OriginLongitude:      req.OriginLongitude,
		DestinationLatitude:  req.DestinationLatitude,
		DestinationLongitude: req.DestinationLongitude,
		DistanceKM:           estimate.DistanceKM,
		Status:               domain.StatusRequested,
		EstimatedPrice:       estimate.EstimatedPrice,
		ETAMinutes:           estimate.ETAMinutes,
	}

	priceHistory := &domain.PriceHistory{
		ID:                   uuid.New(),
		RideID:               ride.ID,
		BaseFare:             estimate.BaseFare,
		DistanceFare:         estimate.DistanceFare,
		SurgeMultiplier:      estimate.SurgeMultiplier,
		FinalCalculatedPrice: estimate.EstimatedPrice,
		IsFallback:           estimate.IsFallback,
	}

	// TRANSACTIONAL OUTBOX PATTERN: Salva corrida, histórico e evento de outbox na mesma transação atômica
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"ride_id":         ride.ID,
		"passenger_id":    ride.PassengerID,
		"estimated_price": ride.EstimatedPrice,
		"status":          ride.Status,
		"timestamp":       time.Now().UTC(),
	})

	outboxEvent := &domain.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "RIDE",
		AggregateID:   ride.ID,
		RoutingKey:    "ride.requested",
		Payload:       eventPayload,
		Status:        domain.OutboxStatusPending,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.rideRepo.CreateRideWithOutbox(ctx, ride, priceHistory, outboxEvent); err != nil {
		return nil, fmt.Errorf("erro ao salvar corrida e evento outbox no banco de dados: %w", err)
	}

	return ride, nil
}

func (s *rideService) AcceptRide(ctx context.Context, rideID uuid.UUID, driverID uuid.UUID) (*domain.Ride, error) {
	driver, err := s.userRepo.GetByID(ctx, driverID)
	if err != nil {
		return nil, fmt.Errorf("motorista não encontrado: %w", err)
	}
	if driver.UserType != domain.UserTypeDriver {
		return nil, errors.New("apenas usuários do tipo 'driver' podem aceitar corridas")
	}

	ride, err := s.rideRepo.GetByID(ctx, rideID)
	if err != nil {
		return nil, err
	}

	if ride.Status != domain.StatusRequested {
		return nil, errors.New("corrida não está disponível para aceite")
	}

	eventPayload, _ := json.Marshal(map[string]interface{}{
		"ride_id":   ride.ID,
		"driver_id": driverID,
		"status":    domain.StatusAccepted,
		"timestamp": time.Now().UTC(),
	})

	outboxEvent := &domain.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "RIDE",
		AggregateID:   ride.ID,
		RoutingKey:    "ride.accepted",
		Payload:       eventPayload,
		Status:        domain.OutboxStatusPending,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.rideRepo.UpdateStatusWithOutbox(ctx, rideID, domain.StatusAccepted, &driverID, nil, outboxEvent); err != nil {
		return nil, fmt.Errorf("erro ao aceitar corrida com outbox: %w", err)
	}

	ride.Status = domain.StatusAccepted
	ride.DriverID = &driverID

	return ride, nil
}

func (s *rideService) CompleteRide(ctx context.Context, rideID uuid.UUID) (*domain.Ride, error) {
	ride, err := s.rideRepo.GetByID(ctx, rideID)
	if err != nil {
		return nil, err
	}

	if ride.Status != domain.StatusAccepted && ride.Status != domain.StatusInProgress {
		return nil, errors.New("apenas corridas aceitas ou em andamento podem ser finalizadas")
	}

	finalPrice := ride.EstimatedPrice
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"ride_id":     ride.ID,
		"final_price": finalPrice,
		"status":      domain.StatusCompleted,
		"timestamp":   time.Now().UTC(),
	})

	outboxEvent := &domain.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "RIDE",
		AggregateID:   ride.ID,
		RoutingKey:    "ride.completed",
		Payload:       eventPayload,
		Status:        domain.OutboxStatusPending,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.rideRepo.UpdateStatusWithOutbox(ctx, rideID, domain.StatusCompleted, nil, &finalPrice, outboxEvent); err != nil {
		return nil, fmt.Errorf("erro ao finalizar corrida com outbox: %w", err)
	}

	ride.Status = domain.StatusCompleted
	ride.FinalPrice = &finalPrice

	return ride, nil
}

func (s *rideService) GetRideDetails(ctx context.Context, rideID uuid.UUID) (*RideDetailsResponse, error) {
	ride, err := s.rideRepo.GetByID(ctx, rideID)
	if err != nil {
		return nil, err
	}

	history, err := s.rideRepo.GetPriceHistoryByRideID(ctx, rideID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar histórico de preços: %w", err)
	}

	return &RideDetailsResponse{
		Ride:         ride,
		PriceHistory: history,
	}, nil
}
