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
	GetArrivals(ctx context.Context, params models.GetStopArrivalsParams) ([]models.Arrival, error)
}

type stopService struct {
	arrivalRepo repositories.ArrivalRepository
	stopRepo    repositories.StopRepository
	logger      *slog.Logger
}

func NewStopService(stopRepo repositories.StopRepository, arrivalRepo repositories.ArrivalRepository, logger *slog.Logger) StopService {
	return &stopService{stopRepo: stopRepo, arrivalRepo: arrivalRepo, logger: logger}
}

func (s *stopService) GetStopByID(ctx context.Context, stopID string) (models.Stop, error) {
	stop, err := s.stopRepo.GetStopByID(ctx, stopID)
	if err != nil {
		return models.Stop{}, fmt.Errorf("service get stop by id: %w", err)
	}
	return stop, nil
}

func (s *stopService) GetArrivals(ctx context.Context, params models.GetStopArrivalsParams) ([]models.Arrival, error) {
	params.NormalizeMinutes()

	arrivals, err := s.arrivalRepo.GetByStop(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("service get arrivals: %w", err)
	}
	return arrivals, nil
}
