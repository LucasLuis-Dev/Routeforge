package service

import (
	"context"
	"errors"
	"testing"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// MockRideRepository
type MockRideRepository struct {
	mock.Mock
}

func (m *MockRideRepository) Create(ctx context.Context, ride *domain.Ride) error {
	args := m.Called(ctx, ride)
	return args.Error(0)
}

func (m *MockRideRepository) CreateRideWithOutbox(ctx context.Context, ride *domain.Ride, history *domain.PriceHistory, outbox *domain.OutboxEvent) error {
	args := m.Called(ctx, ride, history, outbox)
	return args.Error(0)
}

func (m *MockRideRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ride, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Ride), args.Error(1)
}

func (m *MockRideRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RideStatus, driverID *uuid.UUID, finalPrice *float64) error {
	args := m.Called(ctx, id, status, driverID, finalPrice)
	return args.Error(0)
}

func (m *MockRideRepository) UpdateStatusWithOutbox(ctx context.Context, id uuid.UUID, status domain.RideStatus, driverID *uuid.UUID, finalPrice *float64, outbox *domain.OutboxEvent) error {
	args := m.Called(ctx, id, status, driverID, finalPrice, outbox)
	return args.Error(0)
}

func (m *MockRideRepository) SavePriceHistory(ctx context.Context, history *domain.PriceHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockRideRepository) GetPriceHistoryByRideID(ctx context.Context, rideID uuid.UUID) ([]*domain.PriceHistory, error) {
	args := m.Called(ctx, rideID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.PriceHistory), args.Error(1)
}

// MockPredictionClient
type MockPredictionClient struct {
	mock.Mock
}

func (m *MockPredictionClient) Predict(ctx context.Context, req *domain.PredictionRequest) (*domain.PredictionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PredictionResponse), args.Error(1)
}

func (m *MockPredictionClient) Close() error {
	return nil
}


func TestCreateEstimate_Success_WithML(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRideRepo := new(MockRideRepository)
	mockMLClient := new(MockPredictionClient)

	svc := NewRideService(mockUserRepo, mockRideRepo, mockMLClient, nil, nil)

	req := EstimateRequest{
		OriginLatitude:       -23.550520,
		OriginLongitude:      -46.633308,
		DestinationLatitude:  -23.561684,
		DestinationLongitude: -46.655981,
	}

	mockMLClient.On("Predict", mock.Anything, mock.Anything).Return(&domain.PredictionResponse{
		ETAMinutes:      12,
		SurgeMultiplier: 1.30,
		BaseFare:        2.50,
		DistanceFare:    4.64,
		EstimatedPrice:  9.28,
	}, nil)

	resp, err := svc.CreateEstimate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.IsFallback)
	assert.Equal(t, 12, resp.ETAMinutes)
	assert.Equal(t, 1.30, resp.SurgeMultiplier)
	assert.Equal(t, 9.28, resp.EstimatedPrice)
	mockMLClient.AssertExpectations(t)
}

func TestCreateEstimate_Fallback_WhenMLError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRideRepo := new(MockRideRepository)
	mockMLClient := new(MockPredictionClient)

	svc := NewRideService(mockUserRepo, mockRideRepo, mockMLClient, nil, nil)

	req := EstimateRequest{
		OriginLatitude:       -23.550520,
		OriginLongitude:      -46.633308,
		DestinationLatitude:  -23.561684,
		DestinationLongitude: -46.655981,
	}

	// Simula erro de conexão/timeout com o serviço de ML
	mockMLClient.On("Predict", mock.Anything, mock.Anything).Return(nil, errors.New("timeout na conexão"))

	resp, err := svc.CreateEstimate(context.Background(), req)

	assert.NoError(t, err) // Não deve falhar! Deve acionar o fallback
	assert.NotNil(t, resp)
	assert.True(t, resp.IsFallback, "Deve marcar IsFallback como true")
	assert.Equal(t, 1.00, resp.SurgeMultiplier, "Em fallback, surge multiplier é 1.00")
	assert.True(t, resp.EstimatedPrice > 0)
	mockMLClient.AssertExpectations(t)
}

func TestCreateRide_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRideRepo := new(MockRideRepository)
	mockMLClient := new(MockPredictionClient)

	svc := NewRideService(mockUserRepo, mockRideRepo, mockMLClient, nil, nil)

	passengerID := uuid.New()
	passenger := &domain.User{
		ID:       passengerID,
		Name:     "João Silva",
		Email:    "joao@example.com",
		UserType: domain.UserTypePassenger,
	}

	req := CreateRideRequest{
		PassengerID:          passengerID,
		OriginLatitude:       -23.550520,
		OriginLongitude:      -46.633308,
		DestinationLatitude:  -23.561684,
		DestinationLongitude: -46.655981,
	}

	mockUserRepo.On("GetByID", mock.Anything, passengerID).Return(passenger, nil)
	mockMLClient.On("Predict", mock.Anything, mock.Anything).Return(&domain.PredictionResponse{
		ETAMinutes:      10,
		SurgeMultiplier: 1.00,
		BaseFare:        2.50,
		DistanceFare:    4.00,
		EstimatedPrice:  6.50,
	}, nil)

	mockRideRepo.On("CreateRideWithOutbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ride, err := svc.CreateRide(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, ride)
	assert.Equal(t, domain.StatusRequested, ride.Status)
	assert.Equal(t, passengerID, ride.PassengerID)
	assert.Equal(t, 6.50, ride.EstimatedPrice)

	mockUserRepo.AssertExpectations(t)
	mockRideRepo.AssertExpectations(t)
}

func TestAcceptRide_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRideRepo := new(MockRideRepository)
	mockMLClient := new(MockPredictionClient)

	svc := NewRideService(mockUserRepo, mockRideRepo, mockMLClient, nil, nil)

	rideID := uuid.New()
	driverID := uuid.New()

	driver := &domain.User{
		ID:       driverID,
		Name:     "Carlos Motorista",
		Email:    "carlos@driver.com",
		UserType: domain.UserTypeDriver,
	}

	existingRide := &domain.Ride{
		ID:          rideID,
		PassengerID: uuid.New(),
		Status:      domain.StatusRequested,
	}

	mockUserRepo.On("GetByID", mock.Anything, driverID).Return(driver, nil)
	mockRideRepo.On("GetByID", mock.Anything, rideID).Return(existingRide, nil)
	mockRideRepo.On("UpdateStatusWithOutbox", mock.Anything, rideID, domain.StatusAccepted, &driverID, (*float64)(nil), mock.Anything).Return(nil)

	ride, err := svc.AcceptRide(context.Background(), rideID, driverID)

	assert.NoError(t, err)
	assert.NotNil(t, ride)
	assert.Equal(t, domain.StatusAccepted, ride.Status)
	assert.Equal(t, &driverID, ride.DriverID)
}

func TestAcceptRide_InvalidUserType(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRideRepo := new(MockRideRepository)
	mockMLClient := new(MockPredictionClient)

	svc := NewRideService(mockUserRepo, mockRideRepo, mockMLClient, nil, nil)

	rideID := uuid.New()
	passengerID := uuid.New()

	// Tentativa de passageiro aceitar corrida como motorista
	userAsPassenger := &domain.User{
		ID:       passengerID,
		Name:     "Maria Passageira",
		UserType: domain.UserTypePassenger,
	}

	mockUserRepo.On("GetByID", mock.Anything, passengerID).Return(userAsPassenger, nil)

	ride, err := svc.AcceptRide(context.Background(), rideID, passengerID)

	assert.Error(t, err)
	assert.Nil(t, ride)
	assert.Contains(t, err.Error(), "apenas usuários do tipo 'driver' podem aceitar corridas")
}
