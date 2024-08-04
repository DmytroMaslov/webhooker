package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"webhooker/internal/services/models"
)

// look at https://gist.github.com/Ananto30/8af841f250e89c07e122e2a838698246

func (h *Handlers) StreamEvents(w http.ResponseWriter, r *http.Request) {
	orderId := r.PathValue("order_id")
	if orderId == "" {
		http.Error(w, "order_id can't be empty", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	eventCh, done, errCh := h.stream.GetEventStream(r.Context(), orderId)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	for {
		select {
		case event := <-eventCh:
			eventResp := eventToEventResp(event)
			jsonData, _ := json.Marshal(eventResp)
			log.Printf(">>> StreamEvents. [%s] data: %s\n\n", event.OrderStatus, jsonData)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
		case <-done:
			log.Printf("(!) StreamEvents. close connection\n")
			return
		case err := <-errCh:
			log.Printf("(!) StreamEvents. failed to get events stream, err: %s\n", err.Error())
			http.Error(w, "failed to get events stream", http.StatusInternalServerError)
			return
		}
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
