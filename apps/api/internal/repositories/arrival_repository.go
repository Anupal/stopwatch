package repositories

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Anupal/stopwatch/internal/models"
)

type ArrivalRepository interface {
	GetByStop(ctx context.Context, params models.GetStopArrivalsParams) ([]models.Arrival, error)
}

type arrivalRepository struct {
	db           *pgxpool.Pool
	logger       *slog.Logger
	feedTimeZone string
}

func NewArrivalRepository(db *pgxpool.Pool, l *slog.Logger, feedTimeZone string) ArrivalRepository {
	return &arrivalRepository{db: db, logger: l, feedTimeZone: feedTimeZone}
}

func (r *arrivalRepository) GetByStop(ctx context.Context, params models.GetStopArrivalsParams) ([]models.Arrival, error) {
	query := `SELECT * FROM get_upcoming_arrivals($1, $2, $3, $4)`

	r.logger.Debug(
		"Executing arrivals query",
		"query", query,
		"stopID", params.StopID,
		"minutesAhead", params.MinutesAhead,
		"minutesBehind", params.MinutesBehind,
		"timeZone", r.feedTimeZone,
	)
	rows, err := r.db.Query(ctx, query, params.StopID, params.MinutesAhead, params.MinutesBehind, r.feedTimeZone)
	if err != nil {
		return nil, fmt.Errorf("repository query arrivals: %w", err)
	}
	defer rows.Close()

	arrivals, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Arrival])
	if err != nil {
		return nil, fmt.Errorf("repository collect arrivals: %w", err)
	}

	return arrivals, nil
}
