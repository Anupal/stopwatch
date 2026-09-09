package repositories

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Anupal/stopwatch/internal/models"
)

type StopRepository interface {
	GetStopByID(ctx context.Context, stopID string) (models.Stop, error)
}

type stopRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewStopRepository(db *pgxpool.Pool, l *slog.Logger) StopRepository {
	return &stopRepository{db: db, logger: l}
}

func (r *stopRepository) GetStopByID(ctx context.Context, stopID string) (models.Stop, error) {
	query := `
		SELECT stop_id, stop_code, stop_name, stop_lat, stop_lon
		FROM stops
		WHERE stop_id = $1
	`
	var stop models.Stop

	r.logger.Debug(
		"Executing stop by id query",
		"query", query,
		"stopID", stopID,
	)

	err := r.db.QueryRow(ctx, query, stopID).Scan(
		&stop.StopID,
		&stop.StopCode,
		&stop.StopName,
		&stop.StopLat,
		&stop.StopLon,
	)

	if err != nil {
		return models.Stop{}, fmt.Errorf("repository query stop by id: %w", err)
	}

	return stop, nil
}
