package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedEstimate struct {
	DistanceKM      float64 `json:"distance_km"`
	ETAMinutes      int     `json:"eta_minutes"`
	BaseFare        float64 `json:"base_fare"`
	DistanceFare    float64 `json:"distance_fare"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
	EstimatedPrice  float64 `json:"estimated_price"`
	IsFallback      bool    `json:"is_fallback"`
}

type EstimateCache interface {
	GetEstimate(ctx context.Context, origLat, origLng, destLat, destLng float64) (*CachedEstimate, error)
	SetEstimate(ctx context.Context, origLat, origLng, destLat, destLng float64, est *CachedEstimate, ttl time.Duration) error
}

type redisEstimateCache struct {
	rdb *redis.Client
}

func NewEstimateCache(rdb *redis.Client) EstimateCache {
	return &redisEstimateCache{rdb: rdb}
}

func makeCacheKey(origLat, origLng, destLat, destLng float64) string {
	// Arredonda as coordenadas para 3 casas decimais (~100 metros de precisão para cache inteligente)
	return fmt.Sprintf("estimate:%.3f,%.3f->%.3f,%.3f", origLat, origLng, destLat, destLng)
}

func (c *redisEstimateCache) GetEstimate(ctx context.Context, origLat, origLng, destLat, destLng float64) (*CachedEstimate, error) {
	key := makeCacheKey(origLat, origLng, destLat, destLng)
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache MISS
		}
		return nil, err
	}

	var est CachedEstimate
	if err := json.Unmarshal([]byte(val), &est); err != nil {
		return nil, err
	}

	return &est, nil // Cache HIT
}

func (c *redisEstimateCache) SetEstimate(ctx context.Context, origLat, origLng, destLat, destLng float64, est *CachedEstimate, ttl time.Duration) error {
	key := makeCacheKey(origLat, origLng, destLat, destLng)
	data, err := json.Marshal(est)
	if err != nil {
		return err
	}

	return c.rdb.Set(ctx, key, data, ttl).Err()
}
