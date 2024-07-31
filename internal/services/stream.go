package services

import (
	"errors"
	"webhooker/internal/services/models"
	"webhooker/internal/storage/posgres"
)

const (
	defaultLimit     = 10
	defaultOffset    = 0
	defaultSortBy    = models.CreateAt
	defaultSortOrder = models.SortDesc
)

type StreamService struct {
	eventStorage *posgres.EventStorage
	orderStorage *posgres.OrderStorage
}

func NewStreamService(event *posgres.EventStorage, order *posgres.OrderStorage) *StreamService {
	return &StreamService{
		eventStorage: event,
		orderStorage: order,
	}
}

func (s *StreamService) SaveEvent(event *models.Event) error {
	err := s.eventStorage.SaveEvent(event)
	if err != nil {
		return err
	}
	return nil
}

var (
	ErrFilterStatus    = errors.New("provide isFinal or Status")
	ErrOnlyOneRequired = errors.New("only isFinal or Status required")
)

func (s *StreamService) GetOrders(filter *models.OrderFilter) ([]*models.Order, error) {
	// TODO: validate statuses
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
