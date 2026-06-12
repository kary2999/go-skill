// 骨架：HTTP handler 测试（用 httptest）
// 用法：复制到 internal/service/<name>_test.go，替换 TODO。
//
// 核心模式：
//   - httptest.NewRequest 造请求
//   - httptest.NewRecorder 抓响应
//   - mock 下游 usecase

package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHANDLER_METHOD(t *testing.T) { // TODO: TestOrderHandler_Create
	tests := []struct {
		name       string
		body       any
		setupMock  func(uc *mocks.MockTARGETUsecase)
		wantStatus int
		wantJSON   string // 部分匹配用 JSONContains，精确用 JSONEq
	}{
		{
			name: "happy_path_returns_201",
			body: map[string]any{
				// TODO: 填入参
			},
			setupMock: func(uc *mocks.MockTARGETUsecase) {
				// TODO: uc.EXPECT().Create(...).Return(&Order{ID: 1}, nil)
			},
			wantStatus: http.StatusCreated,
			wantJSON:   `{"id":1}`,
		},
		{
			name: "malformed_json_returns_400",
			body: "not-json",
			setupMock: func(uc *mocks.MockTARGETUsecase) {
				// 不应调下游
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "downstream_error_returns_500",
			body: map[string]any{},
			setupMock: func(uc *mocks.MockTARGETUsecase) {
				// uc.EXPECT().Create(...).Return(nil, errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := mocks.NewMockTARGETUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(uc)
			}

			// TODO: h := NewOrderHandler(uc)
			var handler http.HandlerFunc

			var body []byte
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				var err error
				body, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.Background())

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantJSON != "" {
				assert.JSONEq(t, tt.wantJSON, rec.Body.String())
			}
		})
	}
}
