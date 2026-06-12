# 表驱动测试完整套路

## 基础结构

```go
func TestOrderUsecase_Create(t *testing.T) {
    tests := []struct {
        name      string                                  // 场景名（必填）
        setupMock func(repo *mocks.MockOrderRepo)         // mock 准备
        req       *CreateOrderReq                         // 入参
        want      *Order                                  // 期望输出（nil 表示不检查）
        wantErr   error                                   // 期望错误（用 errors.Is 判）
    }{
        {name: "happy path", ...},
        {name: "duplicate key returns existing", ...},
        {name: "invalid amount rejected", wantErr: ErrInvalidArgument},
    }

    for _, tt := range tests {
        tt := tt  // 闭包捕获陷阱，Go 1.22+ 可省略
        t.Run(tt.name, func(t *testing.T) {
            // 每个 case 独立 controller
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            repo := mocks.NewMockOrderRepo(ctrl)
            if tt.setupMock != nil {
                tt.setupMock(repo)
            }

            uc := NewOrderUsecase(repo)
            got, err := uc.Create(context.Background(), tt.req)

            // 错误先判
            if tt.wantErr != nil {
                require.Error(t, err)
                assert.True(t, errors.Is(err, tt.wantErr),
                    "want %v got %v", tt.wantErr, err)
                return
            }
            require.NoError(t, err)

            // 正常结果判
            if tt.want != nil {
                assert.Equal(t, tt.want.ID, got.ID)
                assert.True(t, tt.want.Amount.Equal(got.Amount))
            }
        })
    }
}
```

## 并行测试

```go
t.Run(tt.name, func(t *testing.T) {
    t.Parallel()  // 子测试并行
    // ... 不能共享全局状态
})
```

**注意**：Go 1.21 及以下要 `tt := tt` 捕获；Go 1.22+ 已修复 loop var 问题，可省略。

## 子测试分组

复杂场景可嵌套：

```go
func TestOrderUsecase(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        // Create 的所有 case
    })
    t.Run("Cancel", func(t *testing.T) {
        // Cancel 的所有 case
    })
}
```

跑法：`go test -run 'TestOrderUsecase/Create'`

## 跳过 case

```go
{name: "skip on ci", ...}
```

```go
if tt.name == "skip on ci" && os.Getenv("CI") != "" {
    t.Skip("CI 环境跳过")
}
```

## 常见误区

- ❌ 不用 `t.Run` —— 挂了不知道哪个 case
- ❌ 共享 `ctrl` / `mock` 在循环外 —— 第一个 case 的 EXPECT 影响后面
- ❌ 忘记 `defer ctrl.Finish()` —— 未预期调用不会被捕获
- ❌ 不用 `require` —— nil pointer 后面还继续跑 → panic 替代清晰断言
