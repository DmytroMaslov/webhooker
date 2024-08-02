package services

import (
	"fmt"
	"log"
	"sort"
	"time"
	"webhooker/internal/queue/inmemory"
	"webhooker/internal/schedule/delay"
	"webhooker/internal/services/models"
	"webhooker/internal/storage/posgres"
)

const (
	defaultLimit     = 10
	defaultOffset    = 0
	defaultSortBy    = models.CreateAt
	defaultSortOrder = models.SortDesc
	cooldownTime     = 30 * time.Second
)

type StreamService struct {
	eventStorage *posgres.EventStorage
	orderStorage *posgres.OrderStorage
	broker       *inmemory.Broker
	delay        *delay.Delay
}

func NewStreamService(event *posgres.EventStorage, order *posgres.OrderStorage, broker *inmemory.Broker, delay *delay.Delay) *StreamService {
	return &StreamService{
		eventStorage: event,
		orderStorage: order,
		broker:       broker,
		delay:        delay,
	}
}

func (s *StreamService) SaveEvent(event *models.Event) error {
	if _, ok := models.ValidStatuses[event.OrderStatus]; !ok {
		return fmt.Errorf("invalid status")
	}

	order, err := s.orderStorage.GetOrder(event.OrderID)
	if err != nil {
		return err
	}
	if order.IsFinal {
		return models.ErrAfterFinal
	}

	events, err := s.eventStorage.GetEvents(&models.EventsFilter{OrderID: &event.OrderID})
	if err != nil {
		return err
	}

	for _, e := range events {
		if e.EventID == event.EventID {
			return models.ErrAlreadyExist
		}
	}

	if event.OrderStatus == models.ReturnStatus || event.OrderStatus == models.FailedStatus {
		for _, e := range events {
			if e.OrderStatus == models.DoneStatus {
				return models.ErrAfterFinal
			}
		}
		event.IsFinal = true
	}

	if event.OrderStatus == models.DoneStatus {
		s.changeStatusToFinal(event.OrderID)
	}

	if event.OrderStatus == models.RefundStatus {

		doneEvent := searchEventByStatus(events, models.DoneStatus)

		if doneEvent != nil && event.UpdateAt.Sub(doneEvent.UpdateAt) < cooldownTime {
			event.IsFinal = true
			s.delay.Cancel(event.OrderID)
		} else {
			return models.ErrAfterFinal
		}
	}

	// publish message in queue
	s.broker.Publish(event.OrderID, event)

	// save event and order in db
	s.process(event, order)
	return nil
}

func (s *StreamService) process(event *models.Event, order *models.Order) error {
	err := s.eventStorage.SaveEvent(event)
	if err != nil {
		return fmt.Errorf("failed to process err %w", err)
	}

	// save new order
	if order.ID == "" {
		err := s.orderStorage.SaveOrder(&models.Order{
			ID:       event.OrderID,
			UserID:   event.UserID,
			Status:   event.OrderStatus,
			IsFinal:  event.IsFinal,
			CreateAt: event.CreateAt,
			UpdateAt: event.UpdateAt,
		})
		if err != nil {
			return err
		}
	} else {
		// update existing one
		err := s.orderStorage.UpdateOrder(&models.Order{
			ID:       event.OrderID,
			UserID:   event.UserID,
			Status:   event.OrderStatus,
			IsFinal:  event.IsFinal,
			CreateAt: event.CreateAt,
			UpdateAt: event.UpdateAt,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StreamService) changeStatusToFinal(orderID string) {
	fn := func() {
		order, err := s.orderStorage.GetOrder(orderID)
		if err != nil {
			log.Printf("failed to change order status after cooldown, orderID %s, err:%s", orderID, err.Error())
			return
		}
		if order.IsFinal {
			return
		}
		order.IsFinal = true
		err = s.orderStorage.UpdateOrder(order)
		if err != nil {
			log.Printf("failed to update order after cooldown, orderID %s, err:%s", orderID, err.Error())
		}
		log.Printf("change order_id: %s to final", orderID)
	}

	go s.delay.AddJobFn(orderID, fn, cooldownTime)
}

func searchEventByStatus(events []*models.Event, status string) *models.Event {
	for _, event := range events {
		if event.OrderStatus == status {
			return event
		}
	}
	return nil
}

func (s *StreamService) GetEventStream(orderId string) (*EventStream, error) {
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
