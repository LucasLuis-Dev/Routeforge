package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/google/uuid"
)

type rideRepository struct {
	db *sql.DB
}

func NewRideRepository(db *sql.DB) domain.RideRepository {
	return &rideRepository{db: db}
}

func (r *rideRepository) Create(ctx context.Context, ride *domain.Ride) error {
	if ride.ID == uuid.Nil {
		ride.ID = uuid.New()
	}
	now := time.Now().UTC()
	ride.CreatedAt = now
	ride.UpdatedAt = now

	query := `
		INSERT INTO rides (
			id, passenger_id, driver_id, origin_latitude, origin_longitude,
			destination_latitude, destination_longitude, distance_km, status,
			estimated_price, final_price, eta_minutes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		ride.ID, ride.PassengerID, ride.DriverID, ride.OriginLatitude, ride.OriginLongitude,
		ride.DestinationLatitude, ride.DestinationLongitude, ride.DistanceKM, ride.Status,
		ride.EstimatedPrice, ride.FinalPrice, ride.ETAMinutes, ride.CreatedAt, ride.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("erro ao inserir corrida: %w", err)
	}

	return nil
}

func (r *rideRepository) CreateRideWithOutbox(ctx context.Context, ride *domain.Ride, history *domain.PriceHistory, outbox *domain.OutboxEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	if ride.ID == uuid.Nil {
		ride.ID = uuid.New()
	}
	now := time.Now().UTC()
	ride.CreatedAt = now
	ride.UpdatedAt = now

	queryRide := `
		INSERT INTO rides (
			id, passenger_id, driver_id, origin_latitude, origin_longitude,
			destination_latitude, destination_longitude, distance_km, status,
			estimated_price, final_price, eta_minutes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err = tx.ExecContext(
		ctx, queryRide,
		ride.ID, ride.PassengerID, ride.DriverID, ride.OriginLatitude, ride.OriginLongitude,
		ride.DestinationLatitude, ride.DestinationLongitude, ride.DistanceKM, ride.Status,
		ride.EstimatedPrice, ride.FinalPrice, ride.ETAMinutes, ride.CreatedAt, ride.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("erro ao inserir corrida na transação: %w", err)
	}

	if history != nil {
		if history.ID == uuid.Nil {
			history.ID = uuid.New()
		}
		history.RideID = ride.ID
		if history.CalculatedAt.IsZero() {
			history.CalculatedAt = now
		}

		queryHist := `
			INSERT INTO price_history (
				id, ride_id, base_fare, distance_fare, surge_multiplier,
				final_calculated_price, is_fallback, calculated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8
			)
		`
		_, err = tx.ExecContext(
			ctx, queryHist,
			history.ID, history.RideID, history.BaseFare, history.DistanceFare, history.SurgeMultiplier,
			history.FinalCalculatedPrice, history.IsFallback, history.CalculatedAt,
		)
		if err != nil {
			return fmt.Errorf("erro ao inserir histórico de preço na transação: %w", err)
		}
	}

	if outbox != nil {
		if err := SaveOutboxEventTx(ctx, tx, outbox); err != nil {
			return fmt.Errorf("erro ao salvar evento de outbox na transação: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação da corrida: %w", err)
	}

	return nil
}

func (r *rideRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ride, error) {
	query := `
		SELECT 
			id, passenger_id, driver_id, origin_latitude, origin_longitude,
			destination_latitude, destination_longitude, distance_km, status,
			estimated_price, final_price, eta_minutes, created_at, updated_at
		FROM rides
		WHERE id = $1
	`

	ride := &domain.Ride{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ride.ID, &ride.PassengerID, &ride.DriverID, &ride.OriginLatitude, &ride.OriginLongitude,
		&ride.DestinationLatitude, &ride.DestinationLongitude, &ride.DistanceKM, &ride.Status,
		&ride.EstimatedPrice, &ride.FinalPrice, &ride.ETAMinutes, &ride.CreatedAt, &ride.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRideNotFound
		}
		return nil, fmt.Errorf("erro ao buscar corrida por ID: %w", err)
	}

	return ride, nil
}

func (r *rideRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RideStatus, driverID *uuid.UUID, finalPrice *float64) error {
	now := time.Now().UTC()
	query := `
		UPDATE rides
		SET status = $1,
		    driver_id = COALESCE($2, driver_id),
		    final_price = COALESCE($3, final_price),
		    updated_at = $4
		WHERE id = $5
	`

	res, err := r.db.ExecContext(ctx, query, status, driverID, finalPrice, now, id)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status da corrida: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrRideNotFound
	}

	return nil
}

func (r *rideRepository) UpdateStatusWithOutbox(ctx context.Context, id uuid.UUID, status domain.RideStatus, driverID *uuid.UUID, finalPrice *float64, outbox *domain.OutboxEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação de atualização: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	query := `
		UPDATE rides
		SET status = $1,
		    driver_id = COALESCE($2, driver_id),
		    final_price = COALESCE($3, final_price),
		    updated_at = $4
		WHERE id = $5
	`

	res, err := tx.ExecContext(ctx, query, status, driverID, finalPrice, now, id)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status da corrida na transação: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrRideNotFound
	}

	if outbox != nil {
		if err := SaveOutboxEventTx(ctx, tx, outbox); err != nil {
			return fmt.Errorf("erro ao salvar evento de outbox na atualização: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de atualização: %w", err)
	}

	return nil
}

func (r *rideRepository) SavePriceHistory(ctx context.Context, history *domain.PriceHistory) error {
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}
	if history.CalculatedAt.IsZero() {
		history.CalculatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO price_history (
			id, ride_id, base_fare, distance_fare, surge_multiplier,
			final_calculated_price, is_fallback, calculated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		history.ID, history.RideID, history.BaseFare, history.DistanceFare, history.SurgeMultiplier,
		history.FinalCalculatedPrice, history.IsFallback, history.CalculatedAt,
	)

	if err != nil {
		return fmt.Errorf("erro ao salvar histórico de preço: %w", err)
	}

	return nil
}

func (r *rideRepository) GetPriceHistoryByRideID(ctx context.Context, rideID uuid.UUID) ([]*domain.PriceHistory, error) {
	query := `
		SELECT 
			id, ride_id, base_fare, distance_fare, surge_multiplier,
			final_calculated_price, is_fallback, calculated_at
		FROM price_history
		WHERE ride_id = $1
		ORDER BY calculated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, rideID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar histórico de preço: %w", err)
	}
	defer rows.Close()

	var histories []*domain.PriceHistory
	for rows.Next() {
		h := &domain.PriceHistory{}
		if err := rows.Scan(
			&h.ID, &h.RideID, &h.BaseFare, &h.DistanceFare, &h.SurgeMultiplier,
			&h.FinalCalculatedPrice, &h.IsFallback, &h.CalculatedAt,
		); err != nil {
			return nil, err
		}
		histories = append(histories, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return histories, nil
}
