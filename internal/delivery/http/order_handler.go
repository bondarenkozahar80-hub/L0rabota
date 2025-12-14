package http

import (
	"encoding/json"
	"net/http"
	"order-service0/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type OrderHandler struct {
	orderUseCase usecase.OrderUseCase
}

func NewOrderHandler(orderUseCase usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{
		orderUseCase: orderUseCase,
	}
}

func (h *OrderHandler) GetOrderByUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["id"]

	if orderUID == "" {
		http.Error(w, `{"error": "Order ID is required"}`, http.StatusBadRequest)
		return
	}

	order, err := h.orderUseCase.GetOrderByUID(r.Context(), orderUID)
	if err != nil {
		if errors.Is(err, errors.New("order not found")) {
			http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

func (h *OrderHandler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/index.html")
}
