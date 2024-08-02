package models

import "time"

const (
	OrderCreatedStatus = "cool_order_created"
	PendingStatus      = "sbu_verification_pending"
	ConfirmedStatus    = "confirmed_by_mayor"
	ReturnStatus       = "changed_my_mind"
	FailedStatus       = "failed"
	DoneStatus         = "chinazes"
	RefundStatus       = "give_my_money_back"
)

var ValidStatuses = map[string]bool{
	OrderCreatedStatus: true,
	PendingStatus:      true,
	ConfirmedStatus:    true,
	ReturnStatus:       true,
	FailedStatus:       true,
	DoneStatus:         true,
	RefundStatus:       true,
}

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
