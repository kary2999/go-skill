# 单元测试反模式清单

## 1. 直接调 `time.Now()` / `rand.Int()` / `uuid.New()`

**病**：测试不可重放，同一代码今天过明天挂。

**治**：注入 Clock / Randomizer / IDGenerator。

```go
// 生产代码
type Clock interface{ Now() time.Time }

type OrderUsecase struct {
    clock Clock
}
func (u *OrderUsecase) Create(...) { t := u.clock.Now() }

// 测试
type fakeClock struct{ t time.Time }
func (f *fakeClock) Now() time.Time { return f.t }

uc := &OrderUsecase{clock: &fakeClock{t: mustParse("2026-04-23T10:00:00Z")}}
```

## 2. 共享全局状态

**病**：case A 设置了全局变量，case B 读到 A 的残留，顺序一变就炸。

**治**：每个 case `t.Cleanup(func(){ resetGlobal() })`；能不用全局就不用。

## 3. 真发 HTTP / 真连 DB

**病**：CI 没网就挂；DB schema 变了全坏；慢。

**治**：
- HTTP → `httptest.Server` 或 mock 接口
- DB → `sqlmock` 或集成测试（build tag）
- MQ → mock producer/consumer 接口

## 4. `time.Sleep` 等异步

**病**：慢 + flaky。

**治**：
```go
// 用 channel
done := make(chan struct{})
go func() { defer close(done); doWork() }()
select {
case <-done:
case <-time.After(time.Second):
    t.Fatal("timeout")
}

// 或用 eventually
assert.Eventually(t, func() bool { return state == "done" }, time.Second, 10*time.Millisecond)
```

## 5. 断言字符串匹配错误

**病**：错误文案改一个字就炸。

**治**：`errors.Is` / `errors.As`（见 assertions.md）。

## 6. 一个测试函数测十件事

**病**：失败报一处，真正病灶是别处。

**治**：按"场景/期望结果"拆 case 到表里。

## 7. `assert.Equal(t, expected, got)` 参数顺序反

**病**：报错信息里 "expected" 变 "got"，debug 反向。

**治**：记住 **testify 约定：want 在前，got 在后**。

## 8. `go test` 不带 `-race`

**病**：并发 bug 只在生产裸奔。

**治**：CI 必须跑 `go test -race ./...`。

## 9. 没有 happy path + 失败 path 双覆盖

**病**：只测成功路径，错误处理是盲区。

**治**：每个业务函数至少一组：
- happy path
- 1 个参数校验失败
- 1 个下游依赖返回错误

## 10. Mock 期望设置"全通配"

```go
// ❌ 什么都不验
repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil)

// ✅ 关键参数精确
repo.EXPECT().Create(gomock.Any(), amountGTE{decimal.NewFromInt(100)}).Return(...)
```

## 11. `os.Setenv` 不 restore

**病**：本次 case 改的环境变量留给下一个 case / 下一个测试文件。

**治**：
```go
t.Setenv("FEATURE_X", "1")  // Go 1.17+，自动在 Cleanup 里 restore
```

## 12. `TestMain` 里做全局初始化却不清理

**病**：多个 package 跑 TestMain，资源泄露或端口冲突。

**治**：能放子测试就别放 TestMain；必须放也要 defer 清理。
