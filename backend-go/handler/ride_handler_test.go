package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRideService struct {
	mock.Mock
}

func (m *MockRideService) CreateEstimate(ctx context.Context, req service.EstimateRequest) (*service.EstimateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.EstimateResponse), args.Error(1)
}

func (m *MockRideService) CreateRide(ctx context.Context, req service.CreateRideRequest) (*domain.Ride, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Ride), args.Error(1)
}

func (m *MockRideService) AcceptRide(ctx context.Context, rideID uuid.UUID, driverID uuid.UUID) (*domain.Ride, error) {
	args := m.Called(ctx, rideID, driverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Ride), args.Error(1)
}

func (m *MockRideService) CompleteRide(ctx context.Context, rideID uuid.UUID) (*domain.Ride, error) {
	args := m.Called(ctx, rideID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Ride), args.Error(1)
}

func (m *MockRideService) GetRideDetails(ctx context.Context, rideID uuid.UUID) (*service.RideDetailsResponse, error) {
	args := m.Called(ctx, rideID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RideDetailsResponse), args.Error(1)
}

func TestHealthHandler(t *testing.T) {
	h := NewHealthHandler()

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	h.HealthCheck(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "routeforge-backend")
}

func TestEstimateRideHandler(t *testing.T) {
	mockService := new(MockRideService)
	h := NewRideHandler(mockService)

	reqPayload := service.EstimateRequest{
		OriginLatitude:       -23.550520,
		OriginLongitude:      -46.633308,
		DestinationLatitude:  -23.561684,
		DestinationLongitude: -46.655981,
	}

	mockService.On("CreateEstimate", mock.Anything, reqPayload).Return(&service.EstimateResponse{
		DistanceKM:      2.58,
		ETAMinutes:      10,
		BaseFare:        2.50,
		DistanceFare:    4.64,
		SurgeMultiplier: 1.20,
		EstimatedPrice:  8.57,
		IsFallback:      false,
	}, nil)

	body, _ := json.Marshal(reqPayload)
	req, err := http.NewRequest(http.MethodPost, "/api/v1/rides/estimate", bytes.NewBuffer(body))
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	h.EstimateRide(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "estimated_price")
	mockService.AssertExpectations(t)
}
