package posgres

import (
	"fmt"
	"time"
	"webhooker/config"
	"webhooker/internal/services/models"

	"github.com/lib/pq"
)

type EventRow struct {
	EventID     string
	OrderID     string
	UserID      string
	OrderStatus string
	CreateAt    time.Time
	UpdateAt    time.Time
}

func (e *EventRow) EventRowToEvent() *models.Event {
	return &models.Event{
		EventID:     e.EventID,
		OrderID:     e.OrderID,
		UserID:      e.UserID,
		OrderStatus: e.OrderStatus,
		CreateAt:    e.CreateAt,
		UpdateAt:    e.UpdateAt,
	}
}

func (e *EventRow) EventRowFromEvent(event *models.Event) {
	e.EventID = event.EventID
	e.OrderID = event.OrderID
	e.UserID = event.UserID
	e.OrderStatus = event.OrderStatus
	e.CreateAt = event.CreateAt
	e.UpdateAt = event.UpdateAt
}

type EventStorage struct {
	*PgClient
}

func NewEventStorage(cfg *config.PgCredentials) (*EventStorage, error) {
	client, err := NewPgClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create event storage, err: %w", err)
	}
	return &EventStorage{
		client,
	}, nil
}

func (e *EventStorage) SaveEvent(event *models.Event) error {
	var eventRow EventRow
	eventRow.EventRowFromEvent(event)

	query := "INSERT INTO JustPayEvents(EventID, OrderID, UserID, OrderStatus, CreateAt, UpdateAt) VALUES($1, $2, $3, $4, $5, $6)"

	_, err := e.db.Exec(query, eventRow.EventID, eventRow.OrderID, eventRow.UserID, eventRow.OrderStatus, eventRow.CreateAt, eventRow.UpdateAt)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" {
				return models.ErrAlreadyExist
			}
		}
		return fmt.Errorf("failed to save event, err: %w", err)
	}
	return nil
}

func (e *EventStorage) GetEvents(filter *models.EventsFilter) ([]*models.Event, error) {
	query := `SELECT EventID, OrderID, UserID, OrderStatus, CreateAt, UpdateAt 
	FROM JustPayEvents`

	if filter.OrderID != nil {
		whereStmt := fmt.Sprintf("OrderID = '%s'", *filter.OrderID)
		query = addWhere(query, whereStmt)
	}

	rows, err := e.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events %w", err)
	}

	var events []*models.Event

	for rows.Next() {
		var eventRow EventRow
		err := rows.Scan(&eventRow.EventID, &eventRow.OrderID, &eventRow.UserID, &eventRow.OrderStatus, &eventRow.CreateAt, &eventRow.UpdateAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row %w", err)
		}
		events = append(events, eventRow.EventRowToEvent())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to run get events query: %w", err)
	}

	return events, nil
}
