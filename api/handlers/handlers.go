package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	mux.HandleFunc("GET /orders/{order_id}/events", h.StreamEvents)
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
		if errors.Is(err, models.ErrAfterFinal) {
			http.Error(w, "", http.StatusGone)
			return
		}
		log.Printf("failed to save event: %s", err)
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

// look at https://gist.github.com/Ananto30/8af841f250e89c07e122e2a838698246

func (h *Handlers) StreamEvents(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("order_id")
	if orderId == "" {
		http.Error(w, "order_id can't be empty", http.StatusBadRequest)
		return
	}

	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	stream, err := h.stream.GetEventStream(orderId)
	if err != nil {
		http.Error(w, "failed get stream", http.StatusInternalServerError)
		return
	}

	eventCh := stream.Stream()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	for event := range eventCh {
		eventResp := eventToEventResp(event)
		jsonData, _ := json.Marshal(eventResp)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		w.(http.Flusher).Flush()
	}
}

type EventResp struct {
	OrderId  string `json:"order_id"`
	UserId   string `json:"user_id"`
	Status   string `json:"order_status"`
	IsFinal  bool   `json:"is_final"`
	CreateAt string `json:"created_at"`
	UpdateAt string `json:"updated_at"`
}

func eventToEventResp(e *models.Event) EventResp {
	return EventResp{
		OrderId:  e.OrderID,
		UserId:   e.UserID,
		Status:   e.OrderStatus,
		IsFinal:  e.IsFinal,
		CreateAt: e.CreateAt.Format(timeLayout),
		UpdateAt: e.UpdateAt.Format(timeLayout),
	}
}
