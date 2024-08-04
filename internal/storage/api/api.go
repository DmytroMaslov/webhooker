package api

import "webhooker/internal/services/models"

//go:generate mockgen -source=api.go -destination=mocks/api_mock.go
type OrderStorage interface {
	GetOrder(string) (*models.Order, error)
	GetOrders(*models.OrderFilter) ([]*models.Order, error)
	SaveOrder(*models.Order) error
	UpdateOrder(*models.Order) error
}
