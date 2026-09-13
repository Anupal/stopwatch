package repositories

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Anupal/stopwatch/internal/models"
)

type StopRepository interface {
	GetStopByID(ctx context.Context, stopID string) ([]models.StopAgencyMV, error)
	GetStops(ctx context.Context, agencyIDs []string) ([]models.StopAgencyMV, error)
}

type stopRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewStopRepository(db *pgxpool.Pool, l *slog.Logger) StopRepository {
	return &stopRepository{db: db, logger: l}
}

func (r *stopRepository) GetStopByID(ctx context.Context, stopID string) ([]models.StopAgencyMV, error) {
	query := `
		SELECT
			stop_id, stop_code, stop_name, stop_lat, stop_lon,
			agency_id, agency_name, agency_url
		FROM mv_agency_stops
		WHERE stop_id = $1
	`

	r.logger.Debug(
		"Executing stop by id query",
		"query", query,
		"stopID", stopID,
	)

	rows, err := r.db.Query(ctx, query, stopID)
	if err != nil {
		return nil, fmt.Errorf("repository query stop by id: %w", err)
	}
	defer rows.Close()

	stops, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.StopAgencyMV])
	if err != nil {
		return nil, fmt.Errorf("repository collect stop by id: %w", err)
	}

	return stops, nil
}

func (r *stopRepository) GetStops(ctx context.Context, agencyIDs []string) ([]models.StopAgencyMV, error) {
	query := `
		SELECT
			stop_id, stop_code, stop_name, stop_lat, stop_lon,
			agency_id, agency_name, agency_url
		FROM mv_agency_stops
	`

	var query_args []any

	if len(agencyIDs) > 0 {
		query += ` WHERE agency_id = ANY($1)`
		query_args = append(query_args, agencyIDs)
	}

	r.logger.Debug(
		"Executing stops query",
		"query", query,
		"agency_ids", agencyIDs,
	)

	rows, err := r.db.Query(ctx, query, query_args...)
	if err != nil {
		return nil, fmt.Errorf("repository query stops: %w", err)
	}
	defer rows.Close()

	stops, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.StopAgencyMV])
	if err != nil {
		return nil, fmt.Errorf("repository collect stops: %w", err)
	}

	return stops, nil
}

// func (r *stopRepository) SearchStopsByName(ctx context.Context, searchName string, maxResults int) ([]models.Stop, error) {
// 	query := `
// 		SELECT stop_id, stop_code, stop_name, stop_lat, stop_lon
// 		FROM stops
// 		WHERE agency_id = $1
// 			AND normalized_name % $2
// 		ORDER BY similarity(normalized_name, $2) DESC
// 		LIMIT $3;
// 	`

// 	r.logger.Debug(
// 		"Executing stops by name query",
// 		"query", query,
// 		"search_name", searchName,
// 		"max_results", maxResults,
// 	)

// 	rows, err := r.db.Query(ctx, query, searchName, maxResults)
// 	if err != nil {
// 		return nil, fmt.Errorf("repository query search stops by name: %w", err)
// 	}
// 	defer rows.Close()

// 	stops, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Stop])
// 	if err != nil {
// 		return nil, fmt.Errorf("repository collect search stops by name: %w", err)
// 	}

// 	return stops, nil
// }

// func (r *stopRepository) GetNearestStops(ctx context.Context) ([]models.Stop, error) {
// 	query := `SELECT stop_id, stop_code, stop_name, stop_lat, stop_lon FROM stops`

// 	r.logger.Debug("Executing stops query", "query", query)

// 	rows, err := r.db.Query(ctx, query)
// 	if err != nil {
// 		return nil, fmt.Errorf("repository query stops: %w", err)
// 	}
// 	defer rows.Close()

// 	stops, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Stop])
// 	if err != nil {
// 		return nil, fmt.Errorf("repository collect stops: %w", err)
// 	}

// 	return stops, nil
// }
