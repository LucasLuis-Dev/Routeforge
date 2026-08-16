package domain

import "context"

type PredictionRequest struct {
	DistanceKM float64 `json:"distance_km"`
	HourOfDay  int     `json:"hour_of_day"`
	DayOfWeek  int     `json:"day_of_week"`
}

type PredictionResponse struct {
	ETAMinutes      int     `json:"eta_minutes"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
	BaseFare        float64 `json:"base_fare"`
	DistanceFare    float64 `json:"distance_fare"`
	EstimatedPrice  float64 `json:"estimated_price"`
}

type PredictionClient interface {
	Predict(ctx context.Context, req *PredictionRequest) (*PredictionResponse, error)
}
