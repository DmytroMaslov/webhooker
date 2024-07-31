package models

import (
	"errors"
	"time"
)

var (
	ErrAlreadyExist = errors.New("row already exist")
)

type Event struct {
	EventID     string
	OrderID     string
	UserID      string
	OrderStatus string
	CreateAt    time.Time
	UpdateAt    time.Time
}

type EventsFilter struct {
	OrderID *string
}
