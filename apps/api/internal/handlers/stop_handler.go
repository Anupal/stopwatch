package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Anupal/stopwatch/internal/services"
)

type StopHandler struct {
	stopService services.StopService
	logger      *slog.Logger
}

func NewStopHandler(s services.StopService, l *slog.Logger) *StopHandler {
	return &StopHandler{stopService: s, logger: l}
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

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
