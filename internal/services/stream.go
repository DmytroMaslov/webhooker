package services

import (
	"fmt"
	"sort"
	"webhooker/internal/queue/inmemory"
	"webhooker/internal/services/models"
	"webhooker/internal/storage/posgres"
)

type StreamService struct {
	eventStorage *posgres.EventStorage
	orderStorage *posgres.OrderStorage
	broker       *inmemory.Broker
}

func NewStreamService(event *posgres.EventStorage, order *posgres.OrderStorage, broker *inmemory.Broker) *StreamService {
	return &StreamService{
		eventStorage: event,
		orderStorage: order,
		broker:       broker,
	}
}

func (s *WebhookService) GetEventStream(orderId string) (*EventStream, error) {
	order, err := s.orderStorage.GetOrder(orderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get order %w", err)
	}

	events, err := s.eventStorage.GetEvents(&models.EventsFilter{OrderID: &orderId})
	if err != nil {
		return nil, fmt.Errorf("failed to get events %w", err)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].UpdateAt.Before(events[j].UpdateAt)
	})
	// set last event as final
	for i, event := range events {
		if event.OrderStatus == order.Status {
			events[i].IsFinal = true
		}
	}

	return &EventStream{
		Order:  order,
		Events: events,
	}, nil
}

type EventStream struct {
	Order  *models.Order
	Events []*models.Event
}

func (es *EventStream) Stream() chan *models.Event {
	eventCh := make(chan *models.Event)

	go func() {
		defer close(eventCh)
		for _, event := range es.Events {
			eventCh <- event
		}
	}()

	return eventCh
}
