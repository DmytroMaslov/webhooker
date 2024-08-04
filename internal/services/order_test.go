package services

import (
	"testing"
	"webhooker/internal/services/models"

	apiMock "webhooker/internal/storage/api/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_GetOrders(t *testing.T) {
	var (
		statusDone   = []string{"chinazes"}
		userID       = "testUserID'"
		isFinal      = true
		limitFive    = 5
		defLimit     = defaultLimit
		offset       = 5
		defOffset    = defaultOffset
		sortBy       = models.UpdateAt
		defSortBy    = models.CreateAt
		sortOrder    = models.SortDesc
		defSortOrder = defaultSortOrder
	)
	order := &models.Order{
		ID: "testID",
	}

	testCases := []struct {
		name    string
		args    *models.OrderFilter
		prepare func(*apiMock.MockOrderStorage)
		exp     []*models.Order
		expErr  error
	}{
		{
			name: "all filters, except isFinal",
			args: &models.OrderFilter{
				Status:    statusDone,
				UserID:    &userID,
				Limit:     &limitFive,
				Offset:    &offset,
				SortBy:    &sortBy,
				SortOrder: &sortOrder,
			},
			prepare: func(m *apiMock.MockOrderStorage) {
				m.EXPECT().GetOrders(&models.OrderFilter{
					Status:    statusDone,
					UserID:    &userID,
					Limit:     &limitFive,
					Offset:    &offset,
					SortBy:    &sortBy,
					SortOrder: &sortOrder,
				}).Return([]*models.Order{order}, nil)
			},
			exp: []*models.Order{order},
		},
		{
			name: "filter isFinal + default value",
			args: &models.OrderFilter{
				IsFinal: &isFinal,
			},
			prepare: func(m *apiMock.MockOrderStorage) {
				m.EXPECT().GetOrders(&models.OrderFilter{
					IsFinal:   &isFinal,
					Limit:     &defLimit,
					Offset:    &defOffset,
					SortBy:    &defSortBy,
					SortOrder: &defSortOrder,
				}).Return([]*models.Order{order}, nil)
			},
			exp: []*models.Order{order},
		},
		{
			name:   "failed. isFinal and status absent",
			args:   &models.OrderFilter{},
			expErr: ErrFilterStatus,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctr := gomock.NewController(t)
			defer ctr.Finish()

			orderStorageMock := apiMock.NewMockOrderStorage(ctr)
			if tc.prepare != nil {
				tc.prepare(orderStorageMock)
			}

			s := &OrderService{
				orderStorage: orderStorageMock,
			}

			res, err := s.GetOrders(tc.args)
			assert.Equal(t, tc.exp, res)
			assert.Equal(t, tc.expErr, err)
		})
	}
}
