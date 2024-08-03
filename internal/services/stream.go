package services

import (
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

func (s *WebhookService) GetEventStream(orderId string) (chan *models.Event, chan bool, chan error) {
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
	if es.order.IsFinal && isReadyForStream(es.order.Status, es.events) {
		sort.Slice(es.events, func(i, j int) bool {
			return es.events[i].UpdateAt.Before(es.events[j].UpdateAt)
		})
		// set last event as "final"
		es.events[len(es.events)-1].IsFinal = true

		for _, ev := range es.events {
			es.eventCh <- ev
		}
		es.doneCh <- true
		return
	}
	es.isActive = true
	// if not we need subscribe to receive events
	fmt.Printf("subscribe %s %s\n", es.clientID, es.order.ID)
	queueCh := es.broker.Subscribe(es.clientID, es.order.ID)
	for queueEvent := range queueCh {
		es.eventCh <- queueEvent
		if queueEvent.OrderStatus == models.DoneStatus {
			es.doneCh <- true
			return
		}
	}
}

var statusMinEventCount = map[string]int{
	models.FailedStatus: 1,
	models.ReturnStatus: 1,
	models.DoneStatus:   4,
	models.RefundStatus: 5,
}

func isReadyForStream(orderStatus string, events []*models.Event) bool {
	return len(events) == statusMinEventCount[orderStatus]
}
