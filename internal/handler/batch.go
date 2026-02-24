package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
	"github.com/ivanjtm/YunoChallenge/internal/router"
)

type BatchHandler struct {
	Router *router.Router
}

func (h *BatchHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req model.BatchRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Failed to parse request body: "+err.Error())
		return
	}

	if len(req.Transactions) == 0 {
		WriteError(w, http.StatusBadRequest, "validation_error", "At least 1 transaction is required")
		return
	}
	if len(req.Transactions) > 500 {
		WriteError(w, http.StatusBadRequest, "validation_error", "Maximum 500 transactions per batch")
		return
	}

	for i, tx := range req.Transactions {
		if tx.ID == "" {
			WriteError(w, http.StatusUnprocessableEntity, "validation_error",
				fmt.Sprintf("transactions[%d].id is required", i))
			return
		}
		if tx.Amount <= 0 {
			WriteError(w, http.StatusUnprocessableEntity, "validation_error",
				fmt.Sprintf("transactions[%d].amount must be positive", i))
			return
		}
		if tx.PaymentMethod == "" {
			WriteError(w, http.StatusUnprocessableEntity, "validation_error",
				fmt.Sprintf("transactions[%d].payment_method is required", i))
			return
		}
	}

	result := h.Router.AnalyzeBatch(req.Transactions, time.Now())
	WriteJSON(w, http.StatusOK, result)
}
