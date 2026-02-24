package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/historical"
	"github.com/ivanjtm/YunoChallenge/internal/model"
	"github.com/ivanjtm/YunoChallenge/internal/router"
)

type HistoricalHandler struct {
	Router *router.Router
}

func (h *HistoricalHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req model.HistoricalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Failed to parse request body: "+err.Error())
		return
	}

	if len(req.Transactions) == 0 {
		WriteError(w, http.StatusBadRequest, "validation_error", "At least 1 transaction is required")
		return
	}

	result := historical.Analyze(req.Transactions, h.Router, time.Now())
	WriteJSON(w, http.StatusOK, result)
}
