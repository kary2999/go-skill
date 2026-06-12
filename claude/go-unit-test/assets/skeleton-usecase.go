// 骨架：usecase / biz 层测试
// 用法：复制到 internal/biz/<name>_test.go，替换 TODO 标记，跑 go test ./...
//
// 依赖：
//   - github.com/stretchr/testify
//   - go.uber.org/mock/gomock
//   - mockgen 已生成 mocks/mock_<repo>.go
//
// 生成 mock：在 <repo>.go 头加：
//   //go:generate mockgen -source=<repo>.go -destination=mocks/mock_<repo>.go -package=mocks

package biz_test // TODO: 改成你的包名 + _test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	// TODO: 替换成你的 import 路径
	// "your.com/project/internal/biz"
	// "your.com/project/internal/biz/mocks"
)

func TestTARGET_METHOD(t *testing.T) { // TODO: TestOrderUsecase_Create
	tests := []struct {
		name      string
		setupMock func(repo *mocks.MockTARGETRepo /* TODO */)
		// TODO: 入参字段
		// req *CreateOrderReq
		// TODO: 期望输出字段
		// want *Order
		wantErr error
	}{
		{
			name: "happy_path",
			setupMock: func(repo *mocks.MockTARGETRepo) {
				// TODO: repo.EXPECT()...
			},
			// req:     &CreateOrderReq{...},
			// want:    &Order{ID: 1},
			wantErr: nil,
		},
		{
			name: "invalid_argument_rejected",
			setupMock: func(repo *mocks.MockTARGETRepo) {
				// 不应调用 mock
			},
			// req:     &CreateOrderReq{},
			wantErr: ErrInvalidArgument, // TODO
		},
		// TODO: 至少补一个"下游依赖失败"的 case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockTARGETRepo(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			// TODO: 构造被测对象
			// uc := biz.NewTARGETUsecase(repo)

			// TODO: 调用被测方法
			// got, err := uc.Method(context.Background(), tt.req)
			var err error
			_ = context.Background()

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr),
					"want %v got %v", tt.wantErr, err)
				return
			}
			require.NoError(t, err)

			// TODO: 断言 got 字段
			// assert.Equal(t, tt.want.ID, got.ID)
		})
	}
}
