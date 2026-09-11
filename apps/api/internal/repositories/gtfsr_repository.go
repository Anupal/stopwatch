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
	UpdateArrivalsWithoutRealtime(arrivals []models.Arrival) ([]models.ArrivalResponse, error)
	UpdateArrivalsWithRealtime(stopID string, arrivals []models.Arrival) ([]models.ArrivalResponse, error)
	GetLatestFeed(ctx context.Context) error
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

func (r *gtfsrRepository) UpdateArrivalsWithoutRealtime(arrivals []models.Arrival) ([]models.ArrivalResponse, error) {
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

		arrival.MinutesRemaining = int(arrivalTime.Sub(now).Minutes())
		arrival.Status = "UNKNOWN"

		if !arrivalTime.Before(now) {
			filteredArrivals = append(filteredArrivals, arrival)
		}
	}

	return filteredArrivals, nil
}

func (r *gtfsrRepository) UpdateArrivalsWithRealtime(stopID string, arrivals []models.Arrival) ([]models.ArrivalResponse, error) {
	now := time.Now()
	filteredArrivals := make([]models.ArrivalResponse, 0, len(arrivals))

	r.logger.Debug("Applying realtime updates to arrivals",
		"stop_id", stopID,
		"arrival_count", len(arrivals),
	)

	for i := range arrivals {
		// create ArrivalResponse instance from Arrival instance
		arrival := models.ArrivalResponse{
			Arrival: arrivals[i],
		}

		r.logger.Debug("Applying updates to arrival", "trip_id", arrival.TripId, "route_short_name", arrival.RouteShortName, "expected_arrival_time", arrival.ArrivalTime)

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
		// excluding to avoid trip with separate trip ids but same arrival time
		if !ok {
			r.logger.Debug("No realtime update for trip", "trip_id", arrival.TripId)
			continue
		}

		r.logger.Debug("Looping through all stop updates for trip", "trip_id", arrival.TripId, "stop_count", len(tripUpdate.StopTimeUpdate))
		stopFound := false
		// go through updates for all stops
		for _, stopUpdate := range tripUpdate.StopTimeUpdate {
			if stopUpdate.GetStopId() != stopID {
				continue
			}

			r.logger.Debug("Found realtime stop update",
				"trip_id", arrival.TripId,
				"stop_id", stopID,
			)

			// add arrival and departure delays
			updatedArrivalTime := r.applyStopTimeUpdate(
				&arrival,
				stopUpdate,
				arrivalTime,
				departureTime,
				now,
			)

			// only include arrival after current time
			if !updatedArrivalTime.Before(now) {
				r.logger.Debug("Including arrival",
					"trip_id", arrival.TripId,
					"route_short_name", arrival.RouteShortName,
					"arrival_time", arrival.ArrivalTime,
					"minutes_remaining", arrival.MinutesRemaining,
				)
				filteredArrivals = append(filteredArrivals, arrival)
			} else {
				r.logger.Debug("Excluding past arrival",
					"trip_id", arrival.TripId,
					"route_short_name", arrival.RouteShortName,
					"arrival_time", arrival.ArrivalTime,
				)
			}

			stopFound = true

			// no need to look at remaining stops
			break
		}

		// If the stop was not found,
		// this can happen when real-time delay predictions are unavailable and only stop-time updates
		// for stops the bus has already passed are present.
		// In this case, we fall back to the measured delay from the most recent stop.
		totalStopsInUpdate := len(tripUpdate.StopTimeUpdate)
		if !stopFound && totalStopsInUpdate > 0 {
			r.logger.Debug("Stop not found in stop updates, applying realtime info for the most recent stop")

			mostRecentStopUpdate := tripUpdate.StopTimeUpdate[totalStopsInUpdate-1]
			updatedArrivalTime := r.applyStopTimeUpdate(
				&arrival,
				mostRecentStopUpdate,
				arrivalTime,
				departureTime,
				now,
			)

			r.logger.Debug("Updated Arrival", "trip_id", arrival.TripId, "route_short_name", arrival.RouteShortName, "updated_arrival_time", arrival.ArrivalTime)

			// only include arrival after current time
			if !updatedArrivalTime.Before(now) {
				r.logger.Debug("Including arrival",
					"trip_id", arrival.TripId,
					"route_short_name", arrival.RouteShortName,
					"arrival_time", arrival.ArrivalTime,
					"minutes_remaining", arrival.MinutesRemaining,
				)
				filteredArrivals = append(filteredArrivals, arrival)
			} else {
				r.logger.Debug("Excluding past arrival",
					"trip_id", arrival.TripId,
					"route_short_name", arrival.RouteShortName,
					"arrival_time", arrival.ArrivalTime,
				)
			}
		}
	}

	return filteredArrivals, nil
}

// Fetch latest feed from GTFSR source as protobuf
func (r *gtfsrRepository) GetLatestFeed(ctx context.Context) error {
	// Only fetch is current feed is older than 1 minute
	if time.Since(r.feedFetchedAt) < time.Minute {
		r.logger.Info("GTFS-R feed is fresh, skipping fetching from source")
		return nil
	}

	r.logger.Info("Fetching latest GTFS-R feed", "url", r.feedURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.feedURL, nil)
	if err != nil {
		r.logger.Error("Failed to get create HTTP request for GTFS-R feed", "url", r.feedURL, "error", err)
		return fmt.Errorf("invalid http request %q: %w", r.feedURL, err)
	}
	req.Header.Set("x-api-key", r.feedApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		r.logger.Error("Failed to get latest GTFS-R feed", "url", r.feedURL, "error", err)
		return fmt.Errorf("failed http request %q: %w", r.feedURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error("Failed to read GTFS-R feed response", "url", r.feedURL, "error", err)
		return fmt.Errorf("invalid GTFS-R feed %q: %w", r.feedURL, err)
	}

	feed := gtfs.FeedMessage{}
	err = proto.Unmarshal(body, &feed)
	if err != nil {
		r.logger.Error("Failed to unmarshal GTFS-R feed response into protobuf", "url", r.feedURL, "error", err)
		return fmt.Errorf("invalid GTFS-R feed protobuf %q: %w", r.feedURL, err)
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

	return nil
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

func (r *gtfsrRepository) applyStopTimeUpdate(
	arrival *models.ArrivalResponse,
	stopUpdate *gtfs.TripUpdate_StopTimeUpdate,
	arrivalTime, departureTime, now time.Time,
) time.Time {
	updatedArrivalTime := arrivalTime

	if stopUpdate.Arrival != nil && stopUpdate.Arrival.Delay != nil {
		delay := stopUpdate.Arrival.GetDelay()
		updatedArrivalTime = addDelay(arrivalTime, delay)

		r.logger.Debug("Applied realtime arrival delay",
			"trip_id", arrival.TripId,
			"stop_id", stopUpdate.GetStopId(),
			"scheduled_time", arrivalTime.Format("15:04:05"),
			"delay_secs", delay,
			"updated_time", updatedArrivalTime.Format("15:04:05"),
		)

		arrival.ArrivalTime = updatedArrivalTime.Format("15:04:05")
		arrival.MinutesRemaining = int(math.Round(
			updatedArrivalTime.Sub(now).Minutes(),
		))
	}

	if stopUpdate.Departure != nil && stopUpdate.Departure.Delay != nil {
		delay := stopUpdate.Departure.GetDelay()
		updatedDepartureTime := addDelay(departureTime, delay)

		r.logger.Debug("Applied realtime departure delay",
			"trip_id", arrival.TripId,
			"stop_id", stopUpdate.GetStopId(),
			"scheduled_time", departureTime.Format("15:04:05"),
			"delay_secs", stopUpdate.Departure.GetDelay(),
			"updated_time", updatedDepartureTime.Format("15:04:05"),
		)

		arrival.DepartureTime = updatedDepartureTime.Format("15:04:05")
	}
	arrival.Status = stopUpdate.GetScheduleRelationship().String()
	r.logger.Debug("Updating schedule status", "status", arrival.Status)

	return updatedArrivalTime
}
