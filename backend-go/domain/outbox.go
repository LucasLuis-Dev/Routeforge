package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusProcessed OutboxStatus = "PROCESSED"
	OutboxStatusFailed    OutboxStatus = "FAILED"
)

type OutboxEvent struct {
	ID            uuid.UUID    `json:"id"`
	AggregateType string       `json:"aggregate_type"`
	AggregateID   uuid.UUID    `json:"aggregate_id"`
	RoutingKey    string       `json:"routing_key"`
	Payload       []byte       `json:"payload"`
	Status        OutboxStatus `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	ProcessedAt   *time.Time   `json:"processed_at,omitempty"`
}

type OutboxRepository interface {
	GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)
	MarkProcessed(ctx context.Context, id uuid.UUID) error
}
