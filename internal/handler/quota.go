package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
	"github.com/ivanjtm/YunoChallenge/internal/quota"
)

type QuotaHandler struct {
	Tracker *quota.Tracker
}

func (h *QuotaHandler) Set(w http.ResponseWriter, r *http.Request) {
	var req model.SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Failed to parse request body: "+err.Error())
		return
	}

	if len(req.ProcessorOverrides) == 0 {
		WriteError(w, http.StatusBadRequest, "validation_error", "At least one processor override is required")
		return
	}

	h.Tracker.SetOverrides(req.ProcessorOverrides)

	WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Simulation state updated. Subsequent /refund and /refund/batch calls will use these constraints.",
		"quotas":  h.Tracker.Status(time.Now()),
	})
}

func (h *QuotaHandler) Reset(w http.ResponseWriter, r *http.Request) {
	h.Tracker.ResetOverrides()

	WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Simulation state reset to defaults.",
		"quotas":  h.Tracker.Status(time.Now()),
	})
}
