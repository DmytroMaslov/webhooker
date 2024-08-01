package services

import (
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
