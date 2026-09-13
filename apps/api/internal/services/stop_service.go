package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Anupal/stopwatch/internal/models"
	"github.com/Anupal/stopwatch/internal/repositories"
)

type StopService interface {
	GetStopByID(ctx context.Context, stopID string) (models.Stop, error)
	GetStops(ctx context.Context, agencyIDs []string) ([]models.Stop, error)
	GetNearestStops(ctx context.Context, params models.GetNearestStopsParams) ([]models.Stop, error)
	GetArrivals(ctx context.Context, params models.GetStopArrivalsParams) ([]models.ArrivalResponse, error)
}

type stopService struct {
	arrivalRepo repositories.ArrivalRepository
	stopRepo    repositories.StopRepository
	gtfsrRepo   repositories.GTFSRRepository
	logger      *slog.Logger
}

func NewStopService(
	stopRepo repositories.StopRepository,
	arrivalRepo repositories.ArrivalRepository,
	gtfsrRepo repositories.GTFSRRepository,
	logger *slog.Logger) StopService {
	return &stopService{
		stopRepo:    stopRepo,
		arrivalRepo: arrivalRepo,
		gtfsrRepo:   gtfsrRepo,
		logger:      logger,
	}
}

func (s *stopService) GetStopByID(ctx context.Context, stopID string) (models.Stop, error) {
	stop, err := s.stopRepo.GetStopByID(ctx, stopID)
	if err != nil {
		return models.Stop{}, fmt.Errorf("service get stop by id: %w", err)
	}
	return convertStopAgencyMVtoStop(stop), nil
}

func (s *stopService) GetStops(ctx context.Context, agencyIDs []string) ([]models.Stop, error) {
	mvStops, err := s.stopRepo.GetStops(ctx, agencyIDs)
	if err != nil {
		return nil, fmt.Errorf("service get stops: %w", err)
	}

	// MV can return mutiple rows for combinations of (stop_id, agency_id)
	// so grouping them by stop_id
	mapStopID := make(map[string][]models.StopAgencyMV)
	for _, mvStop := range mvStops {
		mapStopID[mvStop.StopID] = append(mapStopID[mvStop.StopID], mvStop)
	}

	// convert StopsMV to Stops
	stops := make([]models.Stop, 0, len(mapStopID))
	for _, mvStopGroup := range mapStopID {
		stops = append(stops, convertStopAgencyMVtoStop(mvStopGroup))
	}

	return stops, nil
}

func (s *stopService) GetNearestStops(ctx context.Context, params models.GetNearestStopsParams) ([]models.Stop, error) {
	s.logger.Debug("service---", "params", params)

	mvStops, err := s.stopRepo.GetNearestStops(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("service get nearest stops: %w", err)
	}

	// MV can return mutiple rows for combinations of (stop_id, agency_id)
	// so grouping them by stop_id
	mapStopID := make(map[string][]models.StopAgencyMV)
	for _, mvStop := range mvStops {
		mapStopID[mvStop.StopID] = append(mapStopID[mvStop.StopID], mvStop)
	}

	// convert StopsMV to Stops
	stops := make([]models.Stop, 0, len(mapStopID))
	for _, mvStopGroup := range mapStopID {
		stops = append(stops, convertStopAgencyMVtoStop(mvStopGroup))
	}

	return stops, nil
}

func (s *stopService) GetArrivals(ctx context.Context, params models.GetStopArrivalsParams) ([]models.ArrivalResponse, error) {
	params.NormalizeMinutes()

	arrivals, err := s.arrivalRepo.GetByStop(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("service get arrivals: %w", err)
	}

	var filteredArrivals []models.ArrivalResponse

	// update arrivals based on realtime GTFS-R feed
	if params.UseRealTimeFeed {
		err = s.gtfsrRepo.GetLatestFeed(ctx)
		if err != nil {
			return nil, fmt.Errorf("service update GTFS-R feed: %w", err)
		}
		filteredArrivals, err = s.gtfsrRepo.UpdateArrivalsWithRealtime(params.StopID, arrivals)
	} else {
		filteredArrivals, err = s.gtfsrRepo.UpdateArrivalsWithoutRealtime(arrivals)
	}

	if err != nil {
		return nil, fmt.Errorf("service update arrivals: %w", err)
	}

	return filteredArrivals, nil
}

func convertStopAgencyMVtoStop(mvStopGroup []models.StopAgencyMV) models.Stop {
	if len(mvStopGroup) == 0 {
		return models.Stop{}
	}
	stop := models.Stop{
		StopID:   mvStopGroup[0].StopID,
		StopCode: mvStopGroup[0].StopCode,
		StopName: mvStopGroup[0].StopName,
		StopLat:  mvStopGroup[0].StopLat,
		StopLon:  mvStopGroup[0].StopLon,
		Distance: mvStopGroup[0].Distance,
		Agencies: make([]models.Agency, 0, len(mvStopGroup)),
	}

	for _, s := range mvStopGroup {
		stop.Agencies = append(
			stop.Agencies,
			models.Agency{
				AgencyID:   s.AgencyID,
				AgencyName: s.AgencyName,
				AgencyURL:  s.AgencyURL,
			})
	}

	return stop
}
