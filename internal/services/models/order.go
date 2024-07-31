package models

import "time"

type Order struct {
	ID       string
	UserID   string
	Status   string
	IsFinal  bool
	CreateAt time.Time
	UpdateAt time.Time
}

type SortBy string
type SortOrder string

const (
	CreateAt SortBy = "created_at"
	UpdateAt SortBy = "update_at"

	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type OrderFilter struct {
	Status    []string
	UserID    *string
	Limit     *int
	Offset    *int
	IsFinal   *bool
	SortBy    *SortBy
	SortOrder *SortOrder
}
