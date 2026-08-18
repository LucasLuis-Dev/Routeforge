package client

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	pb "github.com/LucasLuis-Dev/Routeforge/backend-go/proto/prediction"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockGRPCServer struct {
	pb.UnimplementedPredictionServiceServer
}

func (s *mockGRPCServer) PredictPricingAndETA(ctx context.Context, req *pb.PredictionRequest) (*pb.PredictionResponse, error) {
	return &pb.PredictionResponse{
		EtaMinutes:      15,
		SurgeMultiplier: 1.25,
		BaseFare:        2.50,
		DistanceFare:    7.20,
		EstimatedPrice:  11.64,
	}, nil
}

func TestGRPCMLClient_Success(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterPredictionServiceServer(grpcServer, &mockGRPCServer{})

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	client, err := NewGRPCMLClient(lis.Addr().String(), 1*time.Second)
	assert.NoError(t, err)
	defer client.Close()

	req := &domain.PredictionRequest{
		DistanceKM:       4.0,
		HourOfDay:        18,
		DayOfWeek:        1,
		TrafficLevel:     1.5,
		WeatherCondition: "RAIN",
	}

	resp, err := client.Predict(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 15, resp.ETAMinutes)
	assert.Equal(t, 1.25, resp.SurgeMultiplier)
	assert.Equal(t, 11.64, resp.EstimatedPrice)
}
