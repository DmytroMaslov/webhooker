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

func (h *Handlers) GetHandlers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhooks/payments/orders", h.ReceiveWebhook)
	mux.Handle("GET /orders", &TimeCounterMiddleware{h.GetOrders}) // count time only for one endpoint
	mux.HandleFunc("GET /orders/{order_id}/events", h.StreamEvents)

	recoverMux := NewRecoverMiddleware(mux) // recover all endpoints
	return recoverMux
}
