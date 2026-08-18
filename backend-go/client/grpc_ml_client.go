package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	pb "github.com/LucasLuis-Dev/Routeforge/backend-go/proto/prediction"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcMLClient struct {
	conn   *grpc.ClientConn
	client pb.PredictionServiceClient
	cb     *gobreaker.CircuitBreaker
}

func NewGRPCMLClient(target string, timeout time.Duration) (domain.PredictionClient, error) {
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao servidor gRPC ML em %s: %w", target, err)
	}

	client := pb.NewPredictionServiceClient(conn)

	st := gobreaker.Settings{
		Name:        "GRPC-ML-Service-CircuitBreaker",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("[CircuitBreaker gRPC] Estado alterado de %s para %s", from.String(), to.String())
		},
	}

	return &grpcMLClient{
		conn:   conn,
		client: client,
		cb:     gobreaker.NewCircuitBreaker(st),
	}, nil
}

func (c *grpcMLClient) Predict(ctx context.Context, req *domain.PredictionRequest) (*domain.PredictionResponse, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		traffic := req.TrafficLevel
		if traffic <= 0 {
			traffic = 1.0
		}
		weather := req.WeatherCondition
		if weather == "" {
			weather = "CLEAR"
		}

		pbReq := &pb.PredictionRequest{
			DistanceKm:       req.DistanceKM,
			HourOfDay:        int32(req.HourOfDay),
			DayOfWeek:        int32(req.DayOfWeek),
			TrafficLevel:     traffic,
			WeatherCondition: weather,
		}

		pbResp, err := c.client.PredictPricingAndETA(ctx, pbReq)
		if err != nil {
			return nil, fmt.Errorf("falha de comunicação gRPC Protobuf com ML Service: %w", err)
		}

		return &domain.PredictionResponse{
			ETAMinutes:      int(pbResp.EtaMinutes),
			SurgeMultiplier: pbResp.SurgeMultiplier,
			BaseFare:        pbResp.BaseFare,
			DistanceFare:    pbResp.DistanceFare,
			EstimatedPrice:  pbResp.EstimatedPrice,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*domain.PredictionResponse), nil
}

func (c *grpcMLClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
