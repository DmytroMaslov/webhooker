package models

import (
	"errors"
	"time"
)

var (
	ErrAlreadyExist = errors.New("row already exist")
	ErrAfterFinal   = errors.New("order in final state")
)

type Event struct {
	EventID     string
	OrderID     string
	UserID      string
	OrderStatus string
	IsFinal     bool
	CreateAt    time.Time
	UpdateAt    time.Time
}

type EventsFilter struct {
	OrderID *string
	EventID *string
}
