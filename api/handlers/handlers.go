package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"webhooker/internal/services"
	"webhooker/internal/services/models"
)

const (
	timeLayout = time.RFC3339
)

type WebhookEvent struct {
	EventID     string `json:"event_id"`
	OrderID     string `json:"order_id"`
	UserID      string `json:"user_id"`
	OrderStatus string `json:"order_status"`
	UpdateAt    string `json:"updated_at"`
	CreateAt    string `json:"created_at"`
}

type Handlers struct {
	stream *services.StreamService
	order  *services.OrderService
}

func NewHandler(stream *services.StreamService, order *services.OrderService) *Handlers {
	return &Handlers{
		stream: stream,
		order:  order,
	}
}

func (h *Handlers) GetHandlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhooks/payments/orders", h.ReceiveWebhook)
	mux.HandleFunc("GET /orders", h.GetOrders)
	return mux
}

func (h *Handlers) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	var webhookEvent WebhookEvent
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&webhookEvent)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	event, err := webhookToEvent(&webhookEvent)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err = h.stream.SaveEvent(event)
	if err != nil {
		if errors.Is(err, models.ErrAlreadyExist) {
			http.Error(w, "", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func webhookToEvent(w *WebhookEvent) (*models.Event, error) {
	var (
		createAtTime time.Time
		updateAtTime time.Time
		err          error
	)
	if createAtTime, err = time.Parse(timeLayout, w.CreateAt); err != nil {
		return nil, err
	}
	if updateAtTime, err = time.Parse(timeLayout, w.UpdateAt); err != nil {
		return nil, err
	}

	return &models.Event{
		EventID:     w.EventID,
		OrderID:     w.OrderID,
		UserID:      w.UserID,
		OrderStatus: w.OrderStatus,
		CreateAt:    createAtTime,
		UpdateAt:    updateAtTime,
	}, nil
}
