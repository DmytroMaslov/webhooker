package handlers

import (
	"net/http"
	"time"
	"webhooker/internal/services"
)

const (
	timeLayout = time.RFC3339
)

type Handlers struct {
	stream *services.WebhookService
	order  *services.OrderService
}

func NewHandler(stream *services.WebhookService, order *services.OrderService) *Handlers {
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
