package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
	"github.com/ivanjtm/YunoChallenge/internal/router"
)

type RefundHandler struct {
	Router *router.Router
}

func (h *RefundHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req model.SingleRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Failed to parse request body: "+err.Error())
		return
	}

	tx := req.Transaction
	if tx.ID == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "transaction.id is required")
		return
	}
	if tx.Country == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "transaction.country is required")
		return
	}
	if tx.Currency == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "transaction.currency is required")
		return
	}
	if tx.PaymentMethod == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "transaction.payment_method is required")
		return
	}
	if tx.Amount <= 0 {
		WriteError(w, http.StatusBadRequest, "validation_error", "transaction.amount must be positive")
		return
	}
	if tx.Timestamp.IsZero() {
		WriteError(w, http.StatusBadRequest, "validation_error", "transaction.timestamp is required")
		return
	}

	result := h.Router.SelectRoute(tx, time.Now())
	WriteJSON(w, http.StatusOK, result)
}
