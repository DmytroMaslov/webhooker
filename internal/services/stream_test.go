package services

import (
	"testing"
	"time"
	"webhooker/internal/services/models"

	"github.com/stretchr/testify/assert"
)

var (
	orderCreateEvent = &models.Event{
		EventID:     "1",
		OrderID:     "1",
		UserID:      "1",
		OrderStatus: models.OrderCreatedStatus,
		CreateAt:    time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC),
		UpdateAt:    time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC),
	}

	pendingEvent = &models.Event{
		EventID:     "2",
		OrderID:     "1",
		UserID:      "1",
		OrderStatus: models.PendingStatus,
		CreateAt:    time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC),
		UpdateAt:    time.Date(2009, 11, 17, 20, 34, 58, 10, time.UTC), // +10
	}

	confirmedEvent = &models.Event{
		EventID:     "3",
		OrderID:     "1",
		UserID:      "1",
		OrderStatus: models.ConfirmedStatus,
		CreateAt:    time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC),
		UpdateAt:    time.Date(2009, 11, 17, 20, 34, 58, 15, time.UTC), // +5
	}

	DoneEventFinal = &models.Event{
		EventID:     "6",
		OrderID:     "1",
		UserID:      "1",
		IsFinal:     true,
		OrderStatus: models.DoneStatus,
		CreateAt:    time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC),
		UpdateAt:    time.Date(2009, 11, 17, 20, 34, 58, 30, time.UTC), // +5
	}

	DoneEventNotFinal = &models.Event{
		EventID:     "6",
		OrderID:     "1",
		UserID:      "1",
		IsFinal:     false,
		OrderStatus: models.DoneStatus,
		CreateAt:    time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC),
		UpdateAt:    time.Date(2009, 11, 17, 20, 34, 58, 30, time.UTC), // +5
	}

	returnEvent = &models.Event{
		EventID:     "4",
		OrderID:     "1",
		UserID:      "1",
		IsFinal:     true,
		OrderStatus: models.ReturnStatus,
		CreateAt:    time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC),
		UpdateAt:    time.Date(2009, 11, 17, 20, 34, 58, 20, time.UTC), // +5
	}

	refundEvent = &models.Event{
		EventID:     "7",
		OrderID:     "1",
		UserID:      "1",
		OrderStatus: models.RefundStatus,
		IsFinal:     true,
		CreateAt:    time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC),
		UpdateAt:    time.Date(2009, 11, 17, 20, 34, 58, 35, time.UTC), // +5 s confirmedEvent
	}

	FailedEvent = &models.Event{
		EventID:     "5",
		OrderID:     "1",
		UserID:      "1",
		OrderStatus: models.FailedStatus,
		CreateAt:    time.Date(2009, 11, 17, 20, 34, 58, 0, time.UTC),
		UpdateAt:    time.Date(2009, 11, 17, 20, 34, 58, 25, time.UTC), // +5
	}
)

func Test_resolve(t *testing.T) {
	type args struct {
		eventsInMemory  []*models.Event
		lastSendedEvent string
	}

	type exp struct {
		events  []*models.Event
		isFinal bool
	}

	testCases := []struct {
		name string
		args args
		exp  exp
	}{

		{
			name: "create event",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent},
				lastSendedEvent: "",
			},
			exp: exp{
				events:  []*models.Event{orderCreateEvent},
				isFinal: false,
			},
		},

		{
			name: "create, pending",
			args: args{
				eventsInMemory:  []*models.Event{pendingEvent, orderCreateEvent},
				lastSendedEvent: "",
			},
			exp: exp{
				events:  []*models.Event{orderCreateEvent, pendingEvent},
				isFinal: false,
			},
		},

		{
			name: "create, pending, confirm",
			args: args{
				eventsInMemory:  []*models.Event{confirmedEvent, pendingEvent, orderCreateEvent},
				lastSendedEvent: "",
			},
			exp: exp{
				events:  []*models.Event{orderCreateEvent, pendingEvent, confirmedEvent},
				isFinal: false,
			},
		},

		{
			name: "create already send, send pending",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent, pendingEvent},
				lastSendedEvent: models.OrderCreatedStatus,
			},
			exp: exp{
				events:  []*models.Event{pendingEvent},
				isFinal: false,
			},
		},

		{
			name: "create, return",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent, returnEvent},
				lastSendedEvent: "",
			},
			exp: exp{
				events:  []*models.Event{orderCreateEvent, returnEvent},
				isFinal: true,
			},
		},
		{
			name: "create, done not final",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent, DoneEventNotFinal},
				lastSendedEvent: "",
			},
			exp: exp{
				events:  []*models.Event{orderCreateEvent},
				isFinal: false,
			},
		},
		{
			name: "create sended, done not final",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent, DoneEventNotFinal},
				lastSendedEvent: orderCreateEvent.OrderStatus,
			},
			exp: exp{
				events:  []*models.Event{},
				isFinal: false,
			},
		},
		{
			name: "create, pending, confirm done final",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent, pendingEvent, confirmedEvent, DoneEventFinal},
				lastSendedEvent: "",
			},
			exp: exp{
				events:  []*models.Event{orderCreateEvent, pendingEvent, confirmedEvent, DoneEventFinal},
				isFinal: true,
			},
		},
		{
			name: "create, pending, return",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent, pendingEvent, returnEvent},
				lastSendedEvent: "",
			},
			exp: exp{
				events:  []*models.Event{orderCreateEvent, pendingEvent, returnEvent},
				isFinal: true,
			},
		},
		{
			name: "create, pending sended, return",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent, pendingEvent, returnEvent},
				lastSendedEvent: pendingEvent.OrderStatus,
			},
			exp: exp{
				events:  []*models.Event{returnEvent},
				isFinal: true,
			},
		},
		{
			name: "create, pending, confirm, done",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent, pendingEvent, confirmedEvent, DoneEventFinal},
				lastSendedEvent: "",
			},
			exp: exp{
				events:  []*models.Event{orderCreateEvent, pendingEvent, confirmedEvent, DoneEventFinal},
				isFinal: true,
			},
		},
		{
			name: "create, pending, confirm, done, refund",
			args: args{
				eventsInMemory:  []*models.Event{orderCreateEvent, pendingEvent, confirmedEvent, DoneEventNotFinal, refundEvent},
				lastSendedEvent: "",
			},
			exp: exp{
				events:  []*models.Event{orderCreateEvent, pendingEvent, confirmedEvent, DoneEventNotFinal, refundEvent},
				isFinal: true,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := eventResolver{
				events:          tc.args.eventsInMemory,
				lastSendedEvent: tc.args.lastSendedEvent,
			}

			resEvents, resDone := resolver.resolve()
			assert.Equal(t, tc.exp.events, resEvents)
			assert.Equal(t, tc.exp.isFinal, resDone)
		})
	}
}

func Test_searchFinalEvent(t *testing.T) {
	testCases := []struct {
		name string
		arg  []*models.Event
		exp  *models.Event
	}{
		{
			name: "with final",
			arg:  []*models.Event{orderCreateEvent, returnEvent},
			exp:  returnEvent,
		},
		{
			name: "without final",
			arg:  []*models.Event{orderCreateEvent, pendingEvent},
			exp:  nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := searchFinalEvent(tc.arg)
			assert.Equal(t, tc.exp, res)
		})
	}
}

func Test_isReadyForStream(t *testing.T) {
	testCases := []struct {
		name string
		argE []*models.Event
		argS string
		exp  bool
	}{
		{
			name: "not ready",
			argE: []*models.Event{orderCreateEvent, refundEvent},
			exp:  false,
		},
		{
			name: "ready",
			argE: []*models.Event{orderCreateEvent, returnEvent},
			exp:  true,
		},
		{
			name: "not ready with done",
			argE: []*models.Event{orderCreateEvent, pendingEvent, confirmedEvent, DoneEventNotFinal},
			exp:  false,
		},
		{
			name: "ready with done",
			argE: []*models.Event{orderCreateEvent, pendingEvent, confirmedEvent, DoneEventFinal},
			exp:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := isReadyForFinalStream(tc.argE)
			assert.Equal(t, tc.exp, res)
		})
	}
}

func Test_appendEvent(t *testing.T) {
	testCases := []struct {
		name      string
		argEvents []*models.Event
		argEvent  *models.Event
		expEvents []*models.Event
	}{
		{
			name:      "add event",
			argEvents: []*models.Event{orderCreateEvent},
			argEvent:  pendingEvent,
			expEvents: []*models.Event{orderCreateEvent, pendingEvent},
		},
		{
			name:      "update event",
			argEvents: []*models.Event{orderCreateEvent, DoneEventNotFinal},
			argEvent:  DoneEventFinal,
			expEvents: []*models.Event{orderCreateEvent, DoneEventFinal},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolver := eventResolver{events: tc.argEvents}
			resolver.appendEvent(tc.argEvent)
			assert.Equal(t, tc.expEvents, resolver.events)
		})
	}
}
