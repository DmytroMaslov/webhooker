package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
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
}

func NewHandler(stream *services.StreamService) *Handlers {
	return &Handlers{
		stream: stream,
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

type OrderResp struct {
	OrderID  string `json:"order_id"`
	UserID   string `json:"user_id"`
	Status   string `json:"status"`
	IsFinal  bool   `json:"is_final"`
	CreateAt string `json:"created_at"`
	UpdateAt string `json:"updated_at"`
}

func orderToOrderResp(order *models.Order) OrderResp {
	return OrderResp{
		OrderID:  order.ID,
		UserID:   order.UserID,
		Status:   order.Status,
		IsFinal:  order.IsFinal,
		CreateAt: order.CreateAt.Format(timeLayout),
		UpdateAt: order.UpdateAt.Format(timeLayout),
	}
}

func (h *Handlers) GetOrders(w http.ResponseWriter, r *http.Request) {
	// status
	statusStr := r.URL.Query().Get("status")
	var statuses []string
	if statusStr != "" {
		statusStr = strings.ReplaceAll(statusStr, " ", "")
		statuses = strings.Split(statusStr, ",")
	}
	// user_id
	userIdStr := r.URL.Query().Get("user_id")
	var userId *string
	if userIdStr != "" {
		userId = &userIdStr
	}
	// limit
	limitStr := r.URL.Query().Get("limit")
	var limit *int
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		if l < 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = &l
	}
	// offset
	offsetStr := r.URL.Query().Get("offset")
	var offset *int
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		if o < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = &o
	}
	// is_final
	isFinalStr := r.URL.Query().Get("isFinal")
	var isFinal *bool
	if isFinalStr != "" {
		b, err := strconv.ParseBool(isFinalStr)
		if err != nil {
			http.Error(w, "invalid isFinal", http.StatusBadRequest)
			return
		}
		isFinal = &b
	}
	// sort_by
	sortByStr := r.URL.Query().Get("sort_by")
	var sortBy *models.SortBy
	if sortByStr != "" {
		if sortByStr != string(models.CreateAt) && sortByStr != string(models.UpdateAt) {
			http.Error(w, "invalid sort_by", http.StatusBadRequest)
			return
		}
		s := models.SortBy(sortByStr)
		sortBy = &s
	}
	// sort_order
	sortOrderStr := r.URL.Query().Get("sort_order")
	var sortOrder *models.SortOrder
	if sortOrderStr != "" {
		if sortOrderStr != string(models.SortAsc) && sortOrderStr != string(models.SortDesc) {
			http.Error(w, "invalid sort_order", http.StatusBadRequest)
			return
		}
		s := models.SortOrder(sortOrderStr)
		sortOrder = &s
	}

	orders, err := h.stream.GetOrders(&models.OrderFilter{
		Status:    statuses,
		UserID:    userId,
		Limit:     limit,
		Offset:    offset,
		IsFinal:   isFinal,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	})

	if err != nil {
		if errors.Is(err, services.ErrFilterStatus) {
			http.Error(w, "provide isFinal or status", http.StatusBadRequest)
			return
		}
		if errors.Is(err, services.ErrOnlyOneRequired) {
			http.Error(w, "provide only isFinal or only Status", http.StatusBadRequest)
			return
		}
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	OrdersResp := make([]OrderResp, 0, len(orders))
	for _, order := range orders {
		OrdersResp = append(OrdersResp, orderToOrderResp(order))
	}

	json, err := json.Marshal(OrdersResp)
	if err != nil {
		http.Error(w, "failed to marshal user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
