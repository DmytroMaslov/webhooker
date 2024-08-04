package posgres

import (
	"fmt"
	"strings"
	"time"
	"webhooker/internal/services/models"
)

type OrderRow struct {
	OrderID     string
	UserID      string
	OrderStatus string
	IsFinal     bool
	CreateAt    time.Time
	UpdateAt    time.Time
}

func (o *OrderRow) OrderRowToOrder() *models.Order {
	return &models.Order{
		ID:       o.OrderID,
		UserID:   o.UserID,
		Status:   o.OrderStatus,
		IsFinal:  o.IsFinal,
		CreateAt: o.CreateAt,
		UpdateAt: o.UpdateAt,
	}
}

func (o *OrderRow) OrderRowFromOrder(order *models.Order) {
	o.OrderID = order.ID
	o.UserID = order.UserID
	o.OrderStatus = order.Status
	o.IsFinal = order.IsFinal
	o.CreateAt = order.CreateAt
	o.UpdateAt = order.UpdateAt
}

type OrderStorage struct {
	db *PgClient
}

func NewOrderStorage(client *PgClient) *OrderStorage {
	return &OrderStorage{
		db: client,
	}
}

func (o *OrderStorage) GetOrder(id string) (*models.Order, error) {
	query := `SELECT OrderID, UserId, OrderStatus, IsFinal, CreateAt, UpdateAt 
	FROM Orders
	WHERE OrderID = $1`

	rows, err := o.db.client.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query order, err: %w", err)
	}
	defer rows.Close()

	var orderRow OrderRow
	for rows.Next() {
		err := rows.Scan(&orderRow.OrderID, &orderRow.UserID, &orderRow.OrderStatus, &orderRow.IsFinal, &orderRow.CreateAt, &orderRow.UpdateAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order row %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to run get order query: %w", err)
	}
	return orderRow.OrderRowToOrder(), nil
}

func (o *OrderStorage) SaveOrder(order *models.Order) error {
	query := `INSERT INTO Orders (OrderID, UserId, OrderStatus, IsFinal, CreateAt, UpdateAt)
	VALUES ( $1, $2, $3, $4, $5, $6)`

	var orderRow OrderRow
	orderRow.OrderRowFromOrder(order)

	_, err := o.db.client.Exec(query, orderRow.OrderID, orderRow.UserID, orderRow.OrderStatus, orderRow.IsFinal, orderRow.CreateAt, orderRow.UpdateAt)
	if err != nil {
		return fmt.Errorf("failed to insert order, err: %w", err)
	}

	return nil
}

func (o *OrderStorage) UpdateOrder(order *models.Order) error {
	query := `UPDATE Orders
	SET OrderID = $1, UserId = $2, OrderStatus = $3, IsFinal = $4, CreateAt = $5, UpdateAt = $6 
	WHERE OrderID = $1`

	var orderRow OrderRow
	orderRow.OrderRowFromOrder(order)

	_, err := o.db.client.Exec(query, orderRow.OrderID, orderRow.UserID, orderRow.OrderStatus, orderRow.IsFinal, orderRow.CreateAt, orderRow.UpdateAt)
	if err != nil {
		return fmt.Errorf("failed to exec update order, err: %w", err)
	}
	return nil
}

func (o *OrderStorage) GetOrders(filter *models.OrderFilter) ([]*models.Order, error) {
	query := `SELECT OrderID, UserId, OrderStatus, IsFinal, CreateAt, UpdateAt 
	FROM Orders`

	if filter.Status != nil {
		var statuses string
		for i := 0; i < len(filter.Status); i++ {
			if i == 0 {
				statuses = fmt.Sprintf("'%s'", filter.Status[i])
			} else {
				statuses = fmt.Sprintf("%s, '%s'", statuses, filter.Status[i])
			}
		}

		whereStmt := fmt.Sprintf("OrderStatus IN (%s)", statuses)
		query = addWhere(query, whereStmt)
	}

	if filter.UserID != nil {
		whereStmt := fmt.Sprintf("UserID = '%s'", *filter.UserID)
		query = addWhere(query, whereStmt)
	}

	if filter.IsFinal != nil {
		whereStmt := fmt.Sprintf("IsFinal = '%t'", *filter.IsFinal)
		query = addWhere(query, whereStmt)
	}

	if filter.SortBy != nil || filter.SortOrder != nil {
		var (
			by    string
			order string
		)
		if *filter.SortBy == models.UpdateAt {
			by = "UpdateAt"
		} else {
			by = "CreateAt"
		}

		if *filter.SortOrder == models.SortAsc {
			order = "asc"
		} else {
			order = "desc"
		}
		sortStmt := fmt.Sprintf("ORDER BY %s %s", by, order)
		query = fmt.Sprintf("%s %s", query, sortStmt)
	}

	if filter.Limit != nil {
		limitStmt := fmt.Sprintf("Limit %d", *filter.Limit)
		query = fmt.Sprintf("%s %s", query, limitStmt)
	}

	if filter.Offset != nil {
		offsetStmt := fmt.Sprintf("Offset %d", *filter.Offset)
		query = fmt.Sprintf("%s %s", query, offsetStmt)
	}

	rows, err := o.db.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query %s, err: %w", query, err)
	}

	var orders []*models.Order

	for rows.Next() {
		var orderRow OrderRow
		err := rows.Scan(&orderRow.OrderID, &orderRow.UserID, &orderRow.OrderStatus, &orderRow.IsFinal, &orderRow.CreateAt, &orderRow.UpdateAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order row %w", err)
		}
		orders = append(orders, orderRow.OrderRowToOrder())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to run get orders query: %w", err)
	}

	return orders, nil
}

func addWhere(query string, stmt string) string {
	if strings.Contains(query, "WHERE") {
		stmt = fmt.Sprintf("AND %s", stmt)
	} else {
		stmt = fmt.Sprintf("WHERE %s", stmt)
	}
	return fmt.Sprintf("%s %s", query, stmt)
}
