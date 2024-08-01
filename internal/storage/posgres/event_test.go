package posgres

import (
	"fmt"
	"testing"
	"time"
	"webhooker/config"
	"webhooker/internal/services/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SaveEvent(t *testing.T) {
	cfg := &config.PgCredentials{
		User:     "test",
		Password: "test",
		Host:     "localhost",
		DbName:   "stream_data",
	}

	s, err := NewEventStorage(cfg)
	require.Nil(t, err)

	event := &models.Event{
		EventID:     "87a96c29-7631-4cbc-9559-f8866fb03392",
		UserID:      "2c127d70-3b9b-4743-9c2e-74b9f617029f",
		OrderID:     "2c127d70-3b9b-4743-9c2e-74b9f617029f",
		OrderStatus: "test",
		CreateAt:    time.Now(),
		UpdateAt:    time.Now(),
	}

	err = s.SaveEvent(event)
	require.Nil(t, err)

	got, err := s.GetEvents(&models.EventsFilter{
		OrderID: &event.OrderID,
	})
	require.Nil(t, err)
	assert.Len(t, got, 1)
	assert.NotEmpty(t, *got[0])

}

func Test_Orders(t *testing.T) {
	cfg := &config.PgCredentials{
		User:     "test",
		Password: "test",
		Host:     "localhost",
		DbName:   "stream_data",
	}

	s, err := NewOrderStorage(cfg)
	require.Nil(t, err)

	order := &models.Order{
		ID:       "27a96c29-7631-4cbc-9559-f8866fb03392",
		UserID:   "2c127d70-3b9b-4743-9c2e-74b9f617029f",
		Status:   "cool_order_created",
		IsFinal:  false,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}

	err = s.SaveOrder(order)
	require.Nil(t, err)

	order.IsFinal = true

	err = s.UpdateOrder(order)
	require.Nil(t, err)

	get, err := s.GetOrder(order.ID)
	require.Nil(t, err)
	fmt.Printf("get: %+v\n", get)
}

func Test_Orders_Get(t *testing.T) {
	cfg := &config.PgCredentials{
		User:     "test",
		Password: "test",
		Host:     "localhost",
		DbName:   "stream_data",
	}

	s, err := NewOrderStorage(cfg)
	require.Nil(t, err)

	//userId := "3c127d70-3b9b-4743-9c2e-74b9f617029f"
	limit := 1
	offset := 1
	//isFinal := true
	sortBy := models.CreateAt
	sortOrder := models.SortAsc

	filter := &models.OrderFilter{
		//Status:    []string{"cool_order_created"},
		//UserID:    &userId,
		Limit:  &limit,
		Offset: &offset,
		//IsFinal:   &isFinal,
		SortBy:    &sortBy,
		SortOrder: &sortOrder,
	}
	res, err := s.GetOrders(filter)
	fmt.Printf("err: %s\n", err)
	for i, o := range res {
		fmt.Printf("res %d: %+v\n", i, o)
	}

	t.Fail()
}
