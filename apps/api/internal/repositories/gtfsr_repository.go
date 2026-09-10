package repositories

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/Anupal/stopwatch/internal/models"
	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

type GTFSRRepository interface {
	UpdateArrivalsWithRealtime(ctx context.Context, stopID string, arrivals []models.Arrival) ([]models.ArrivalResponse, error)
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

func (r *gtfsrRepository) UpdateArrivalsWithRealtime(ctx context.Context, stopID string, arrivals []models.Arrival) ([]models.ArrivalResponse, error) {
	r.getLatestFeed(ctx)

	now := time.Now()
	filteredArrivals := make([]models.ArrivalResponse, 0, len(arrivals))

	for i := range arrivals {
		// create ArrivalResponse instance from Arrival instance
		arrival := models.ArrivalResponse{
			Arrival: arrivals[i],
		}

		// get parsed arrival and departure times
		arrivalTime, err := parseTimeToday(arrival.ArrivalTime, now)
		if err != nil {
			return nil, fmt.Errorf("invalid time %q: %w", arrival.ArrivalTime, err)
		}
		departureTime, err := parseTimeToday(arrival.DepartureTime, now)
		if err != nil {
			return nil, fmt.Errorf("invalid time %q: %w", arrival.DepartureTime, err)
		}

		tripUpdate, ok := r.tripUpdateMap[arrival.TripId]

		// no realtime data for this trip.
		if !ok {
			// --- note: logic to add scheduled arrival unchanged                       ---
			// --- excluding to avoid trip with separate trip ids but same arrival time ---
			// arrival.MinutesRemaining = int(arrivalTime.Sub(now).Minutes())
			// arrival.Status = "SCHEDULED"

			// if !arrivalTime.Before(now) {
			// 	filteredArrivals = append(filteredArrivals, arrival)
			// }
			continue
		}

		// go through updates for all stops
		for _, stopUpdate := range tripUpdate.StopTimeUpdate {
			if stopUpdate.GetStopId() != stopID {
				continue
			}

			// add arrival and departure delays
			var updatedArrivalTime, updatedDepartureTime time.Time

			if stopUpdate.Arrival != nil && stopUpdate.Arrival.Delay != nil {
				updatedArrivalTime = addDelay(arrivalTime, stopUpdate.Arrival.GetDelay())
				arrival.ArrivalTime = updatedArrivalTime.Format("15:04:05")

				// computes minutes left until arrival
				minutesLeft := updatedArrivalTime.Sub(now).Minutes()
				arrival.MinutesRemaining = int(math.Round(minutesLeft))
			}
			if stopUpdate.Departure != nil && stopUpdate.Departure.Delay != nil {
				updatedDepartureTime = addDelay(departureTime, stopUpdate.Departure.GetDelay())
				arrival.DepartureTime = updatedDepartureTime.Format("15:04:05")
			}

			arrival.Status = stopUpdate.GetScheduleRelationship().String()

			// only include arrival after current time
			if !updatedArrivalTime.Before(now) {
				filteredArrivals = append(filteredArrivals, arrival)
			}

			// no need to look at remaining stops
			break
		}
	}

	return filteredArrivals, nil
}

// Fetch latest feed from GTFSR source as protobuf
func (r *gtfsrRepository) getLatestFeed(ctx context.Context) {
	// Only fetch is current feed is older than 1 minute
	if time.Since(r.feedFetchedAt) < time.Minute {
		r.logger.Info("GTFS-R feed is fresh, skipping fetching from source")
		return
	}

	r.logger.Info("Fetching latest GTFS-R feed", "url", r.feedURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.feedURL, nil)
	if err != nil {
		r.logger.Error("Failed to get create HTTP request for GTFS-R feed", "url", r.feedURL, "error", err)
		return
	}
	req.Header.Set("x-api-key", r.feedApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		r.logger.Error("Failed to get latest GTFS-R feed", "url", r.feedURL, "error", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("Failed to read GTFS-R feed response", "url", r.feedURL, "error", err)
		return
	}

	feed := gtfs.FeedMessage{}
	err = proto.Unmarshal(body, &feed)
	if err != nil {
		r.logger.Error("Failed to unmarshal GTFS-R feed response into protobuf", "url", r.feedURL, "error", err)
		return
	}

	// todo: fix concurrency issues with simultaneous HTTP requests
	// save map for tripID -> tripUpdate
	r.tripUpdateMap = make(map[string]*gtfs.TripUpdate)

	for _, entity := range feed.Entity {
		tripUpdate := entity.GetTripUpdate()
		if tripUpdate == nil {
			continue
		}

		tripID := tripUpdate.GetTrip().GetTripId()
		r.tripUpdateMap[tripID] = tripUpdate
	}

	// save timestamp
	r.feedFetchedAt = time.Now()
}

func addDelay(scheduled time.Time, delay int32) time.Time {
	return scheduled.Add(time.Duration(delay) * time.Second)
}

func parseTimeToday(timeStr string, now time.Time) (time.Time, error) {
	t, err := time.ParseInLocation("15:04:05", timeStr, now.Location())
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location()), nil
}
