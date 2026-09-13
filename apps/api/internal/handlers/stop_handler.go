package handlers

import (
	"encoding/json"
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

	agencyIDs := r.URL.Query()["agency_id"]

	stops, err := h.stopService.GetStops(r.Context(), agencyIDs)
	if err != nil {
		h.logger.Error("Failed to fetch stops", "agency_ids", agencyIDs, "error", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to fetch stops")
		return
	}

	h.logger.Info("Successfully fetched stops", "agency_ids", agencyIDs, "num_stop", len(stops))
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
