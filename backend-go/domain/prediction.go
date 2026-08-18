package domain

import "context"

type PredictionRequest struct {
	DistanceKM       float64 `json:"distance_km"`
	HourOfDay        int     `json:"hour_of_day"`
	DayOfWeek        int     `json:"day_of_week"`
	TrafficLevel     float64 `json:"traffic_level,omitempty"`     // 1.0 = livre, 1.5 = moderado, 2.0 = pesado
	WeatherCondition string  `json:"weather_condition,omitempty"` // "CLEAR", "RAIN", "STORM"
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
	Close() error
}
