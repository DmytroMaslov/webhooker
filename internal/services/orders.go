package services

import (
	"errors"
	"webhooker/internal/services/models"
	"webhooker/internal/storage/posgres"
)

type OrderService struct {
	orderStorage *posgres.OrderStorage
}

func NewOrderService(order *posgres.OrderStorage) *OrderService {
	return &OrderService{
		orderStorage: order,
	}
}

var (
	ErrFilterStatus      = errors.New("provide isFinal or Status")
	ErrOnlyOneRequired   = errors.New("only isFinal or Status required")
	ErrUnsupportedStatus = errors.New("unsupported status")
)

func (s *OrderService) GetOrders(filter *models.OrderFilter) ([]*models.Order, error) {
	if filter.IsFinal == nil && filter.Status == nil {
		return nil, ErrFilterStatus
	}
	if filter.IsFinal != nil && filter.Status != nil {
		return nil, ErrOnlyOneRequired
	}
	limit := defaultLimit
	if filter.Limit != nil {
		limit = *filter.Limit
	}

	offset := defaultOffset
	if filter.Offset != nil {
		offset = *filter.Offset
	}

	sortBy := defaultSortBy
	if filter.SortBy != nil {
		sortBy = *filter.SortBy
	}

	sortOrder := defaultSortOrder
	if filter.SortOrder != nil {
		sortOrder = *filter.SortOrder
	}

	for _, s := range filter.Status {
		_, ok := models.StatusPriority[s]
		if !ok {
			return nil, ErrUnsupportedStatus
		}
	}

	orderFilter := &models.OrderFilter{
		Status:    filter.Status,
		UserID:    filter.UserID,
		Limit:     &limit,
		Offset:    &offset,
		IsFinal:   filter.IsFinal,
		SortBy:    &sortBy,
		SortOrder: &sortOrder,
	}
	return s.orderStorage.GetOrders(orderFilter)
}
