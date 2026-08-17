package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/google/uuid"
)

type outboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) domain.OutboxRepository {
	return &outboxRepository{db: db}
}

func SaveOutboxEventTx(ctx context.Context, tx *sql.Tx, event *domain.OutboxEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.Status == "" {
		event.Status = domain.OutboxStatusPending
	}

	query := `
		INSERT INTO outbox_events (
			id, aggregate_type, aggregate_id, routing_key, payload, status, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	_, err := tx.ExecContext(
		ctx, query,
		event.ID, event.AggregateType, event.AggregateID, event.RoutingKey, event.Payload, event.Status, event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("erro ao inserir evento na outbox: %w", err)
	}

	return nil
}

func (r *outboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, aggregate_type, aggregate_id, routing_key, payload, status, created_at, processed_at
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar eventos pendentes na outbox: %w", err)
	}
	defer rows.Close()

	var events []*domain.OutboxEvent
	for rows.Next() {
		ev := &domain.OutboxEvent{}
		var procAt sql.NullTime
		if err := rows.Scan(
			&ev.ID, &ev.AggregateType, &ev.AggregateID, &ev.RoutingKey, &ev.Payload, &ev.Status, &ev.CreatedAt, &procAt,
		); err != nil {
			return nil, err
		}
		if procAt.Valid {
			ev.ProcessedAt = &procAt.Time
		}
		events = append(events, ev)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *outboxRepository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	query := `
		UPDATE outbox_events
		SET status = 'PROCESSED', processed_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("erro ao marcar evento como processado na outbox: %w", err)
	}

	return nil
}
