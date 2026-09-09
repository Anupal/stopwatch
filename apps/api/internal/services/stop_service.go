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
}

type stopService struct {
	stopRepo repositories.StopRepository
	logger   *slog.Logger
}

func NewStopService(stopRepo repositories.StopRepository, logger *slog.Logger) StopService {
	return &stopService{stopRepo: stopRepo, logger: logger}
}

func (s *stopService) GetStopByID(ctx context.Context, stopID string) (models.Stop, error) {
	stop, err := s.stopRepo.GetStopByID(ctx, stopID)
	if err != nil {
		return models.Stop{}, fmt.Errorf("service get stop by id: %w", err)
	}
	return stop, nil
}
