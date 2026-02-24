package handler

import (
	"net/http"

	internalconfig "github.com/ivanjtm/YunoChallenge/internal/config"
)

type HealthHandler struct {
	Config *internalconfig.AppConfig
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{
		"status":                 "ok",
		"processors_loaded":      len(h.Config.Processors),
		"rules_loaded":           len(h.Config.Rules),
		"transactions_available": len(h.Config.Transactions),
	})
}
