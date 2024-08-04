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
	IsFinal     bool
	CreateAt    time.Time
	UpdateAt    time.Time
}

func (e *EventRow) EventRowToEvent() *models.Event {
	return &models.Event{
		EventID:     e.EventID,
		OrderID:     e.OrderID,
		UserID:      e.UserID,
		OrderStatus: e.OrderStatus,
		IsFinal:     e.IsFinal,
		CreateAt:    e.CreateAt,
		UpdateAt:    e.UpdateAt,
	}
}

func (e *EventRow) EventRowFromEvent(event *models.Event) {
	e.EventID = event.EventID
	e.OrderID = event.OrderID
	e.UserID = event.UserID
	e.OrderStatus = event.OrderStatus
	e.IsFinal = event.IsFinal
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

	query := "INSERT INTO Events(EventID, OrderID, UserID, OrderStatus, IsFinal, CreateAt, UpdateAt) VALUES($1, $2, $3, $4, $5, $6, $7)"

	_, err := e.db.Exec(query, eventRow.EventID, eventRow.OrderID, eventRow.UserID, eventRow.OrderStatus, eventRow.IsFinal, eventRow.CreateAt, eventRow.UpdateAt)
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

func (e *EventStorage) UpdateEvent(event *models.Event) error {
	query := `UPDATE Events
	SET EventID = $1, OrderID = $2, UserID = $3, OrderStatus = $4, IsFinal = $5, CreateAt = $6, UpdateAt = $7
	WHERE EventID = $1`

	var eventRow EventRow
	eventRow.EventRowFromEvent(event)

	_, err := e.db.Exec(query, eventRow.EventID, eventRow.OrderID, eventRow.UserID, eventRow.OrderStatus, eventRow.IsFinal, eventRow.CreateAt, eventRow.UpdateAt)
	if err != nil {
		return fmt.Errorf("failed to update event, err: %w", err)
	}
	return nil
}

func (e *EventStorage) GetEvents(filter *models.EventsFilter) ([]*models.Event, error) {
	query := `SELECT EventID, OrderID, UserID, OrderStatus, IsFinal, CreateAt, UpdateAt 
	FROM Events`

	if filter.OrderID != nil {
		whereStmt := fmt.Sprintf("OrderID = '%s'", *filter.OrderID)
		query = addWhere(query, whereStmt)
	}

	if filter.EventID != nil {
		whereStmt := fmt.Sprintf("EventID = '%s'", *filter.EventID)
		query = addWhere(query, whereStmt)
	}

	rows, err := e.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events %w", err)
	}

	var events []*models.Event

	for rows.Next() {
		var eventRow EventRow
		err := rows.Scan(&eventRow.EventID, &eventRow.OrderID, &eventRow.UserID, &eventRow.OrderStatus, &eventRow.IsFinal, &eventRow.CreateAt, &eventRow.UpdateAt)
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
