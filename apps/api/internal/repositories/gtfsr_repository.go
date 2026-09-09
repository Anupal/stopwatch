package repositories

import (
	"log/slog"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

type GTFSRRepository interface {
}

type gtfsrRepository struct {
	feedURL    string
	feedApiKey string
	logger     *slog.Logger

	// map tripID -> tripUpdate
	tripUpdateMap map[string]*gtfs.TripUpdate
	feedFetchedAt time.Time
}

func NewGTFSRRepository(feedURL string, feedApiKey string, l *slog.Logger) GTFSRRepository {
	return &gtfsrRepository{logger: l, feedApiKey: feedApiKey, feedURL: feedURL}
}
