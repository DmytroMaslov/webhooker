package models

import "time"

const (
	CooldownTime = 30 * time.Second

	OrderCreatedStatus = "cool_order_created"
	PendingStatus      = "sbu_verification_pending"
	ConfirmedStatus    = "confirmed_by_mayor"
	ReturnStatus       = "changed_my_mind"
	FailedStatus       = "failed"
	DoneStatus         = "chinazes"
	RefundStatus       = "give_my_money_back"
)

var StatusPriority = map[string]int{
	OrderCreatedStatus: 1,
	PendingStatus:      2,
	ConfirmedStatus:    3,
	ReturnStatus:       4,
	FailedStatus:       5,
	DoneStatus:         6,
	RefundStatus:       7,
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
