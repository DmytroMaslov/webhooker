package services

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"
	"webhooker/internal/queue/inmemory"
	"webhooker/internal/services/models"
	"webhooker/internal/storage/posgres"

	"github.com/google/uuid"
)

const (
	waitTime = 60 * time.Second
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

func (s *WebhookService) GetEventStream(ctx context.Context, orderId string) (chan *models.Event, chan bool, chan error) {
	log.Printf("in GetEventStream\n")
	eventsCh := make(chan *models.Event)
	doneCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		defer close(eventsCh)
		defer close(doneCh)
		defer close(errCh)

		order, err := s.orderStorage.GetOrder(orderId)
		if err != nil {
			errCh <- fmt.Errorf("failed to get order %w", err)
			return
		}
		// if we don't have order we need order id for event subscription
		if order.ID == "" {
			order.ID = orderId
		}

		events, err := s.eventStorage.GetEvents(&models.EventsFilter{OrderID: &orderId})
		if err != nil {
			errCh <- fmt.Errorf("failed to get events %w", err)
			return
		}

		es := NewEventStream(order, events, s.broker)
		go es.Stream()
		defer es.CleanUp()

		for {
			select {
			case <-time.After(waitTime):
				log.Printf("time.After %fs. in Stream\n", waitTime.Seconds())
				if !es.isActive {
					doneCh <- true
					return
				}
			case message, ok := <-es.eventCh:
				if ok {
					eventsCh <- message
				}
			case <-es.doneCh:
				doneCh <- true
				return
			case <-ctx.Done():
				doneCh <- true
				return
			}
		}
	}()

	return eventsCh, doneCh, errCh
}

type EventStream struct {
	broker   *inmemory.Broker
	order    *models.Order
	events   []*models.Event
	isActive bool
	clientID string
	eventCh  chan *models.Event
	doneCh   chan bool
}

// maybe move in Stream()
func (es *EventStream) CleanUp() {
	log.Printf("in CleanUp\n")
	close(es.eventCh)
	close(es.doneCh)
	es.broker.UnSubscribe(es.clientID, es.order.ID)
}

func NewEventStream(order *models.Order, events []*models.Event, br *inmemory.Broker) *EventStream {
	log.Printf("in NewEventStream\n")
	return &EventStream{
		broker:   br,
		order:    order,
		events:   events,
		clientID: uuid.NewString(),
		eventCh:  make(chan *models.Event),
		doneCh:   make(chan bool),
	}
}

func (es *EventStream) Stream() {
	log.Printf("in Stream\n")

	// check if we can stream all data from db
	if es.order.IsFinal && isReadyForFinalStream(es.events) {
		for _, ev := range es.events {
			es.eventCh <- ev
		}
		es.doneCh <- true
		return
	}

	// check if we have initial event
	orderCreatedEvent := searchEventByStatus(es.events, models.OrderCreatedStatus)
	if orderCreatedEvent != nil {
		es.isActive = true
	}
	// sent event that ready
	eventResolver := eventResolver{events: es.events}
	eventForStream, _ := eventResolver.resolve()
	for _, e := range eventForStream {
		es.eventCh <- e
	}

	// subscribe
	queueCh := es.broker.Subscribe(es.clientID, es.order.ID)
	for queueEvent := range queueCh {

		// we receive initial event
		if queueEvent.OrderStatus == models.OrderCreatedStatus {
			es.isActive = true
		}
		eventResolver.appendEvent(queueEvent)
		eventForStream, done := eventResolver.resolve()
		for _, e := range eventForStream {
			es.eventCh <- e
		}
		if done {
			es.doneCh <- true
			return
		}
	}
}

type eventResolver struct {
	lastSendedEvent string
	events          []*models.Event
}

func (er *eventResolver) appendEvent(event *models.Event) {
	isNew := true
	for i, e := range er.events {
		if e.EventID == event.EventID {
			er.events[i] = event
			isNew = false
		}
	}
	if isNew {
		er.events = append(er.events, event)
	}
}

func (er *eventResolver) resolve() ([]*models.Event, bool) {
	eventForSending := make([]*models.Event, 0)

	if len(er.events) == 0 {
		return []*models.Event{}, false
	}
	// we have final event and min amount of events
	if isReadyForFinalStream(er.events) {
		var index int
		if er.lastSendedEvent != "" {
			index = searchStatusIndex(er.events, er.lastSendedEvent)
			eventForSending = append(eventForSending, er.events[index+1:]...)
			return eventForSending, true
		} else {
			eventForSending = append(eventForSending, er.events...)
			return eventForSending, true
		}
	}

	sort.Slice(er.events, func(i, j int) bool {
		return er.events[i].UpdateAt.Before(er.events[j].UpdateAt)
	})

	initialEvent := searchEventByStatus(er.events, models.OrderCreatedStatus)
	if initialEvent == nil {
		return []*models.Event{}, false
	}

	eventForSending = append(eventForSending, initialEvent)

	if len(er.events) == 1 {
		if er.lastSendedEvent == "" {
			er.lastSendedEvent = eventForSending[len(eventForSending)-1].OrderStatus
			return eventForSending, false
		} else {
			return []*models.Event{}, false
		}
	}

	for i := 1; i < len(er.events); i++ {
		if models.StatusPriority[er.events[i].OrderStatus]-models.StatusPriority[er.events[i-1].OrderStatus] == 1 {
			eventForSending = append(eventForSending, er.events[i])
		}
	}

	if er.lastSendedEvent != "" {
		i := searchStatusIndex(eventForSending, er.lastSendedEvent)
		er.lastSendedEvent = eventForSending[len(eventForSending)-1].OrderStatus
		return eventForSending[i+1:], false
	}
	er.lastSendedEvent = eventForSending[len(eventForSending)-1].OrderStatus
	return eventForSending, false
}

var statusMinEventCount = map[string]int{
	models.FailedStatus: 2,
	models.ReturnStatus: 2,
	models.DoneStatus:   4,
	models.RefundStatus: 5,
}

// check if we have final event and min event amount by type
func isReadyForFinalStream(events []*models.Event) bool {
	sort.Slice(events, func(i, j int) bool {
		return events[i].UpdateAt.Before(events[j].UpdateAt)
	})
	if len(events) < 2 {
		return false
	}

	finalEvent := searchFinalEvent(events)
	if finalEvent == nil {
		return false
	}

	lastEvent := events[len(events)-1]
	return len(events) >= statusMinEventCount[lastEvent.OrderStatus]
}

func searchEventByStatus(events []*models.Event, status string) *models.Event {
	for _, event := range events {
		if event.OrderStatus == status {
			return event
		}
	}
	return nil
}

func searchFinalEvent(events []*models.Event) *models.Event {
	for _, e := range events {
		if e.IsFinal {
			return e
		}
	}
	return nil
}

func searchStatusIndex(events []*models.Event, status string) int {
	for i, e := range events {
		if e.OrderStatus == status {
			return i
		}
	}
	return 0
}
