// Demo: 表驱动单测 + mockgen — 对应 testing.md 规范
package biz_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOrderUsecase_CreateOrder(t *testing.T) {
	tests := []struct {
		name      string
		req       *CreateOrderReq
		setupMock func(repo *MockOrderRepository, lock *MockDistLocker)
		wantErr   error // 用 errors.Is 断言
	}{
		{
			name: "valid order creates successfully",
			req:  &CreateOrderReq{IdempotencyKey: "idem-1", Amount: decimal.NewFromInt(100)},
			setupMock: func(repo *MockOrderRepository, lock *MockDistLocker) {
				repo.EXPECT().GetByIdempotencyKey(gomock.Any(), "idem-1").Return(nil, ErrNotFound)
				lock.EXPECT().Acquire(gomock.Any(), "order:create:idem-1").Return(func() {}, nil)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&Order{ID: 1}, nil)
			},
			wantErr: nil,
		},
		{
			name: "missing idempotency key rejected",
			req:  &CreateOrderReq{Amount: decimal.NewFromInt(100)},
			setupMock: func(repo *MockOrderRepository, lock *MockDistLocker) {
				// 不应调用任何 mock
			},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "idempotent replay returns existing",
			req:  &CreateOrderReq{IdempotencyKey: "idem-2", Amount: decimal.NewFromInt(100)},
			setupMock: func(repo *MockOrderRepository, lock *MockDistLocker) {
				repo.EXPECT().GetByIdempotencyKey(gomock.Any(), "idem-2").
					Return(&Order{ID: 42, IdempotencyKey: "idem-2"}, nil)
				// 幂等命中后不应再加锁/写入
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := NewMockOrderRepository(ctrl)
			lock := NewMockDistLocker(ctrl)
			tt.setupMock(repo, lock)

			uc := &OrderUsecase{repo: repo, locker: lock}
			_, err := uc.CreateOrder(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr), "want %v got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 生成 mock 的方法（放在被 mock 接口所在文件的 // go:generate 注释）：
//   //go:generate mockgen -source=order_repo.go -destination=mock_order_repo.go -package=biz
//
// 这样 `go generate ./...` 会重建 mock 文件；禁止手写 mock。

// 约定命名：Test{被测对象}_{场景}_{期望结果}
// 对应第三组用例： TestOrderUsecase_CreateOrder_IdempotentReplay_ReturnsExisting
