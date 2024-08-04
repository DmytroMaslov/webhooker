package posgres

import (
	"testing"
	"time"
	"webhooker/internal/services/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var (
	ordersColumn = []string{"OrderID", "UserId", "OrderStatus", "IsFinal", "CreateAt", "UpdateAt"}
)

func Test_GetOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	expOrder := &models.Order{
		ID:       "testID",
		UserID:   "userID",
		Status:   "testStatus",
		IsFinal:  true,
		CreateAt: time.Date(2022, 10, 10, 11, 30, 30, 0, time.UTC),
		UpdateAt: time.Date(2022, 10, 10, 11, 30, 30, 0, time.UTC),
	}
	orderRow := sqlmock.NewRows(ordersColumn).
		AddRow(expOrder.ID, expOrder.UserID, expOrder.Status, expOrder.IsFinal, expOrder.CreateAt, expOrder.UpdateAt)

	mock.ExpectQuery(`SELECT OrderID, UserId, OrderStatus, IsFinal, CreateAt, UpdateAt FROM Orders WHERE OrderID = \$1`).WithArgs(expOrder.ID).WillReturnRows(orderRow)

	storage := OrderStorage{db: &PgClient{db}}

	order, err := storage.GetOrder(expOrder.ID)
	assert.Nil(t, err)
	assert.Equal(t, expOrder, order)
}
