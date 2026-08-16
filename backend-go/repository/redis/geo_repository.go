package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const driversGeoKey = "drivers:locations"

type DriverLocation struct {
	DriverID   string  `json:"driver_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	DistanceKm float64 `json:"distance_km"`
}

type GeoRepository interface {
	UpdateDriverLocation(ctx context.Context, driverID string, lat, lng float64) error
	FindNearbyDrivers(ctx context.Context, lat, lng float64, radiusKm float64, count int) ([]DriverLocation, error)
}

type redisGeoRepository struct {
	rdb *redis.Client
}

func NewGeoRepository(rdb *redis.Client) GeoRepository {
	return &redisGeoRepository{rdb: rdb}
}

func (r *redisGeoRepository) UpdateDriverLocation(ctx context.Context, driverID string, lat, lng float64) error {
	// Redis GEOADD requer Longitude primeiro, depois Latitude
	err := r.rdb.GeoAdd(ctx, driversGeoKey, &redis.GeoLocation{
		Name:      driverID,
		Longitude: lng,
		Latitude:  lat,
	}).Err()

	if err != nil {
		return fmt.Errorf("erro ao atualizar localização GEO no Redis: %w", err)
	}
	return nil
}

func (r *redisGeoRepository) FindNearbyDrivers(ctx context.Context, lat, lng float64, radiusKm float64, count int) ([]DriverLocation, error) {
	if radiusKm <= 0 {
		radiusKm = 5.0 // padrão 5 km
	}
	if count <= 0 {
		count = 5 // padrão 5 motoristas mais próximos
	}

	locations, err := r.rdb.GeoSearchLocation(ctx, driversGeoKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lng,
			Latitude:   lat,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "ASC",
			Count:      count,
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("erro ao buscar motoristas próximos no Redis GEO: %w", err)
	}

	result := make([]DriverLocation, 0, len(locations))
	for _, loc := range locations {
		result = append(result, DriverLocation{
			DriverID:   loc.Name,
			Latitude:   loc.Latitude,
			Longitude:  loc.Longitude,
			DistanceKm: loc.Dist,
		})
	}

	return result, nil
}
