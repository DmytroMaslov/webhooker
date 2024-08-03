package services

import (
	"fmt"
	"log"
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
)

type WebhookService struct {
	eventStorage *posgres.EventStorage
	orderStorage *posgres.OrderStorage
	broker       *inmemory.Broker
	delay        *delay.Delay
}

func NewWebhookService(event *posgres.EventStorage, order *posgres.OrderStorage, broker *inmemory.Broker, delay *delay.Delay) *WebhookService {
	return &WebhookService{
		eventStorage: event,
		orderStorage: order,
		broker:       broker,
		delay:        delay,
	}
}

func (s *WebhookService) SaveEvent(event *models.Event) error {
	if _, ok := models.StatusPriority[event.OrderStatus]; !ok {
		return fmt.Errorf("unsupported status")
	}

	order, err := s.orderStorage.GetOrder(event.OrderID)
	if err != nil {
		return err
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

	if event.OrderStatus == models.PendingStatus || event.OrderStatus == models.ConfirmedStatus {
		doneEvent := searchEventByStatus(events, models.DoneStatus)
		if doneEvent != nil && event.UpdateAt.After(doneEvent.UpdateAt) {
			return models.ErrAfterFinal
		}

		refundEvent := searchEventByStatus(events, models.RefundStatus)
		if refundEvent != nil {
			return models.ErrAfterFinal
		}

		failedEvent := searchEventByStatus(events, models.FailedStatus)
		if failedEvent != nil {
			return models.ErrAfterFinal
		}
	}

	if event.OrderStatus == models.ReturnStatus || event.OrderStatus == models.FailedStatus {
		// check if order in final state
		if order.IsFinal {
			return models.ErrAfterFinal
		}

		// check if order already has done status
		for _, e := range events {
			if e.OrderStatus == models.DoneStatus {
				return models.ErrAfterFinal
			}
		}
		order.IsFinal = true
		order.Status = event.OrderStatus
		order.UpdateAt = event.UpdateAt
	}

	if event.OrderStatus == models.DoneStatus {
		s.changeStatusToFinal(event.OrderID)
	}

	if event.OrderStatus == models.RefundStatus {

		doneEvent := searchEventByStatus(events, models.DoneStatus)

		if doneEvent != nil && event.UpdateAt.Sub(doneEvent.UpdateAt) < models.CooldownTime {
			order.IsFinal = true
			order.Status = event.OrderStatus
			order.UpdateAt = event.UpdateAt

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

func (s *WebhookService) process(event *models.Event, order *models.Order) error {
	err := s.eventStorage.SaveEvent(event)
	if err != nil {
		return fmt.Errorf("failed to process err %w", err)
	}

	// update order only if priority of new event higher than event in order
	// for example we can receive DoneStatus and after than PendingStatus

	if models.StatusPriority[event.OrderStatus] < models.StatusPriority[order.Status] {
		return nil
	}

	// save new order
	if order.ID == "" {
		err := s.orderStorage.SaveOrder(&models.Order{
			ID:       event.OrderID,
			UserID:   event.UserID,
			Status:   event.OrderStatus,
			IsFinal:  order.IsFinal,
			CreateAt: event.CreateAt,
			UpdateAt: event.UpdateAt,
		})
		if err != nil {
			return err
		}
	} else {
		// update existing one
		err := s.orderStorage.UpdateOrder(&models.Order{
			ID:       order.ID,
			UserID:   order.UserID,
			Status:   order.Status,
			IsFinal:  order.IsFinal,
			CreateAt: order.CreateAt,
			UpdateAt: order.UpdateAt,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *WebhookService) changeStatusToFinal(orderID string) {
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

	go s.delay.AddJobFn(orderID, fn, models.CooldownTime)
}

func searchEventByStatus(events []*models.Event, status string) *models.Event {
	for _, event := range events {
		if event.OrderStatus == status {
			return event
		}
	}
	return nil
}
