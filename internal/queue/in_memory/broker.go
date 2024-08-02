package inmemory

import (
	"sync"
	"webhooker/internal/services/models"
)

type Broker struct {
	mu     sync.Mutex
	subs   map[string]map[string]chan *models.Event
	quit   chan struct{}
	closed bool
}

func NewBroker() *Broker {
	return &Broker{
		subs: make(map[string]map[string]chan *models.Event),
		quit: make(chan struct{}),
	}
}

func (b *Broker) Publish(topic string, msg *models.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	for t, clients := range b.subs {
		if t == topic {
			for _, ch := range clients {
				ch <- msg
			}
		}
	}
}

func (b *Broker) Subscribe(clientId string, topic string) chan *models.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	ch := make(chan *models.Event)

	if _, ok := b.subs[topic]; !ok {
		b.subs[topic] = make(map[string]chan *models.Event)
	}

	b.subs[topic][clientId] = ch
	return ch
}

func (b *Broker) UnSubscribe(clientId string, topic string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	delete(b.subs[topic], clientId)
}

func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
	close(b.quit)

	for _, ch := range b.subs {
		for _, sub := range ch {
			close(sub)
		}
	}

	for _, clients := range b.subs {
		for _, ch := range clients {
			close(ch)
		}
	}
}
