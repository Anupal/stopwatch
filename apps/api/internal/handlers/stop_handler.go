package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Anupal/stopwatch/internal/models"
	"github.com/Anupal/stopwatch/internal/services"
)

type StopHandler struct {
	stopService services.StopService
	logger      *slog.Logger
}

func NewStopHandler(s services.StopService, l *slog.Logger) *StopHandler {
	return &StopHandler{stopService: s, logger: l}
}

// GET /stops
func (h *StopHandler) GetStops(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	agencyIDs := r.URL.Query()["agencyId"]

	stops, err := h.stopService.GetStops(r.Context(), agencyIDs)
	if err != nil {
		h.logger.Error("Failed to fetch stops", "agency_ids", agencyIDs, "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to fetch stops")
		return
	}

	h.logger.Info("Successfully fetched stops", "agency_ids", agencyIDs, "num_stop", len(stops))
	json.NewEncoder(w).Encode(stops)
}

// GET /stops/search?query={query}&maxResults={number}
func (h *StopHandler) GetStopsByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	agencyIDs := r.URL.Query()["agencyId"]

	searchQuery := query.Get("query")
	if searchQuery == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing 'query' query param.")
	}

	params := models.GetStopsByNameParams{
		AgencyIDs:   agencyIDs,
		SearchQuery: searchQuery,
		MaxResults:  parseIntParam(query.Get("maxResults"), 5),
	}

	stops, err := h.stopService.GetStopsByName(r.Context(), params)
	if err != nil {
		h.logger.Error("Failed to fetch stops by name", "params", params, "error", err)
		writeJSONError(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf("Failed to fetch stops by name - %v", params),
		)
		return
	}

	h.logger.Info("Successfully fetched stops by name", "agency_ids", agencyIDs, "num_stop", len(stops))
	json.NewEncoder(w).Encode(stops)
}

// GET /stops/nearby?agency={agency1ID}&agency={agency2ID}&latitude={latitude}&longitude={longitude}
func (h *StopHandler) GetNearestStops(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	agencyIDs := query["agencyId"]
	// Note: Disabling to default to return all agencies
	// if len(agencyIDs) == 0 {
	// 	writeJSONError(w, http.StatusBadRequest, "Missing 'agencyId' query param.")
	// 	return
	// }
	latitude := query.Get("latitude")
	if latitude == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing 'latitude' query param.")
		return
	}
	longitude := query.Get("longitude")
	if longitude == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing 'longitude' query param.")
		return
	}
	maxDistance := query.Get("maxDistance")
	if maxDistance == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing 'maxDistance' query param.")
	}

	params := models.GetNearestStopsParams{
		AgencyIDs:       agencyIDs,
		Latitude:        latitude,
		Longitude:       longitude,
		MaximumDistance: maxDistance,
	}

	stops, err := h.stopService.GetNearestStops(r.Context(), params)
	if err != nil {
		h.logger.Error("Failed to fetch nearest stops", "agency_ids", agencyIDs, "error", err)
		writeJSONError(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf("Failed to fetch stops near to (%v, %v)", latitude, longitude),
		)
		return
	}

	h.logger.Info("Successfully fetched nearest stops", "agency_ids", agencyIDs, "num_stop", len(stops))
	json.NewEncoder(w).Encode(stops)
}

// GET /stops/{stopID}
func (h *StopHandler) GetStopByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stopID := r.PathValue("stopID")

	h.logger.Info("Getting details for stop", "stopID", stopID)

	stop, err := h.stopService.GetStopByID(r.Context(), stopID)
	if err != nil {
		h.logger.Error("Failed to fetch details for stop", "stopID", stopID, "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to find stop")
		return
	}
	h.logger.Info("Successfully fetched details for stop", "stopID", stopID, "stop", stop)
	json.NewEncoder(w).Encode(stop)
}

// GET /stops/{stopID}/arrivals
func (h *StopHandler) GetArrivals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	stopID := r.PathValue("stopID")
	expandStop := query.Get("expandStop") == "true"
	useRealTimeFeed := query.Get("realtime") == "true"

	h.logger.Info("Getting arrivals for stop", "stopID", stopID)

	params := models.GetStopArrivalsParams{
		StopID:          stopID,
		MinutesAhead:    parseIntParam(query.Get("minutesAhead"), 0),
		UseRealTimeFeed: useRealTimeFeed,
	}

	arrivals, err := h.stopService.GetArrivals(r.Context(), params)
	if err != nil {
		h.logger.Error("Failed to fetch arrivals for stop", "stopID", stopID, "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to fetch arrivals")
		return
	}

	response := models.GetArrivalsResponse{
		MinutesAhead: params.MinutesAhead,
		Arrivals:     arrivals,
	}

	if expandStop {
		stop, err := h.stopService.GetStopByID(r.Context(), params.StopID)
		if err != nil {
			h.logger.Error("Failed to fetch details for stop", "stopID", stopID, "error", err)
			writeJSONError(w, http.StatusInternalServerError, "Failed to find stop")
			return
		}
		h.logger.Info("Successfully fetched details for stop", "stopID", stopID, "stop", stop)
		response.Stop = &stop

	} else {
		response.StopID = stopID
	}

	h.logger.Info("Successfully fetched arrivals for stop", "stopID", stopID, "arrivals", arrivals)
	json.NewEncoder(w).Encode(response)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func parseIntParam(val string, defaultVal int) int {
	convertedVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}

	return convertedVal
}
