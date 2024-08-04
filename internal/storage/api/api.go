package api

import "webhooker/internal/services/models"

type OrderStorage interface {
	GetOrder(string) (*models.Order, error)
	GetOrders(*models.OrderFilter) ([]*models.Order, error)
	SaveOrder(*models.Order) error
	UpdateOrder(*models.Order) error
}
