package inmemory

import (
	"testing"
	"webhooker/internal/services/models"

	"github.com/stretchr/testify/assert"
)

func Test_Broker(t *testing.T) {
	// Create a new agent
	agent := NewBroker()

	// Subscribe to a topic
	client1 := agent.Subscribe("client1", "order1")
	client2 := agent.Subscribe("client2", "order1")

	// Publish a message to the topic
	go agent.Publish("order1", &models.Event{
		EventID: "event1",
	})

	// Print the message
	eventFromClient1 := <-client1
	eventFromClient2 := <-client2

	agent.UnSubscribe("client1", "order1")
	agent.UnSubscribe("client2", "order1")
	agent.Close()

	assert.Equal(t, "event1", eventFromClient1.EventID)
	assert.Equal(t, "event1", eventFromClient2.EventID)
}
