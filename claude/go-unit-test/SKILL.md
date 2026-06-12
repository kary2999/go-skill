---
name: go-unit-test
description: |
  【Go 单元测试专用 · 强制前置激活】
  只要用户任务涉及写 Go 测试代码，立即读完本 SKILL.md 再动手。
  触发信号（模糊匹配，任一命中即激活）：
  - 动词：写测试 / 补测试 / 加单测 / 加一下 test / test 一下 / 覆盖一下 / 写个 mock / 补覆盖率
  - 文件：*_test.go / testdata/ / mock_*.go / mocks/
  - 工具：gomock / mockgen / testify / go test / -cover / httptest / sqlmock / dockertest
  - 结构：表驱动 / table-driven / subtests / t.Run / golden file / fixture
  - 场景：单元测试 / 集成测试 / unit test / integration test / TDD / 覆盖率 / coverage
  - 被测：usecase / biz / service / repo / handler / middleware 后接"测试/test"
  - 【次级触发】提测 / QA 交付 / 生成提测报告 / OrangeCat —— 与 orangecat 并发激活，验证 Go 单测覆盖是否达标
  违反 = 凭感觉写 assert.Equal + 硬编码数据 = 团队规范破 = 返工。
  与 go-team-standards、orangecat 并列激活，不互斥。
---

# Go Unit Test Skill

写 Go 单元测试时的唯一指引。基础层（本文件）自给自足；深化知识按可选模块装。

## 🚨 ZERO STEP：写测试前必做

1. 识别被测对象类型（usecase / repo / handler / middleware / 工具函数）
2. 按下方「铁律」+「场景 → 文件」路由读对应 references（若已装）
3. 查反模式清单（若 `references/anti-patterns.md` 已装，必读）
4. **能抄骨架就抄**（若 `assets/skeleton-*.go` 已装），不要从 0 敲

## 🔴 铁律 10 条（不依赖任何可选模块，这里就是完整规则）

| # | 规则 | 违反后果 |
|---|---|---|
| 1 | **表驱动**：多 case 必用 `tests := []struct{...}` + `t.Run(tt.name, ...)` | 多 case 堆 if = 可读性崩，失败定位难 |
| 2 | **禁手写 mock**：一律 `mockgen` 生成 + 源文件头加 `//go:generate mockgen ...` | 接口改了忘改 mock = 假绿 |
| 3 | **禁真实外部 IO**：单测不碰真 DB / 真 MQ / 真 HTTP 网络；需要时走集成测（build tag `integration`） | 单测慢 + CI 不稳 + 依赖外部环境 |
| 4 | **错误断言用 `errors.Is` / `errors.As`**：禁 `err.Error() == "xxx"` 或 `strings.Contains(err.Error(), ...)` | 包装链或文案变了就炸 |
| 5 | **时间/随机/ID 必须注入**：禁直接调 `time.Now()` / `rand.Int()` / `uuid.New()` | 测试不可重放，偶发失败 |
| 6 | **金额比较用 `decimal.Equal`**：禁 `assert.Equal` 比 float / decimal | 浮点误差 + decimal 精度字段差异 |
| 7 | **测试独立**：每个 case 不依赖其他 case 的残留状态；清理用 `t.Cleanup()` | 顺序依赖 = 偶发失败 |
| 8 | **命名 `Test{对象}_{方法}_{场景}_{期望}`**，例：`TestOrderUsecase_Create_IdempotentReplay_ReturnsExisting` | 报错看不出测什么 |
| 9 | **覆盖率目标**：核心业务 biz/usecase ≥ 90%，基础工具 ≥ 80%，以 `go test -cover` 为准 | 覆盖不到 = 未验证 |
| 10 | **集成测试用 build tag** `//go:build integration`，单独跑；主 CI 跑 unit | unit 和集成混跑拖慢 CI + 依赖环境 |

## 🧭 场景 → 文件路由（模糊匹配被测对象）

| 被测对象 | 读文件（若已装） | 抄骨架（若已装） |
|---|---|---|
| usecase / biz 业务层 | `references/table-driven.md` + `references/mock-gomock.md` | `assets/skeleton-usecase.go` |
| repo / dao 数据层 | `references/mock-gomock.md` + `references/integration.md` | `assets/skeleton-repo-sqlmock.go` |
| HTTP handler | `references/table-driven.md` + `references/assertions.md` | `assets/skeleton-http-handler.go` |
| middleware | `references/table-driven.md` | `assets/skeleton-http-handler.go` |
| 工具函数 / pure func | `references/table-driven.md` | `assets/skeleton-usecase.go`（简化版） |
| MQ consumer | `references/mock-gomock.md`（mock producer/consumer 接口） | — |
| 涉及时间/随机/ID | `references/anti-patterns.md` | — |

> 💡 文件标注「若已装」—— 本 Skill 采用模块化，可在 Team Standards App 勾选安装。
> 未装时按 SKILL.md 铁律写仍然合规；装了深化可参考。

## 写测试标准步骤（不依赖可选模块也能走通）

### 1. 建文件

```
被测：internal/biz/order.go
测试：internal/biz/order_test.go   （外部包 order_biz_test，避免耦合私有字段）
mock：internal/biz/mocks/mock_order_repo.go
```

源文件头加 `//go:generate`：
```go
//go:generate mockgen -source=order_repo.go -destination=mocks/mock_order_repo.go -package=mocks
```

### 2. 表驱动骨架（没装骨架包时按这个写）

```go
package biz_test

import (
    "context"
    "errors"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/mock/gomock"
)

func TestOrderUsecase_Create(t *testing.T) {
    tests := []struct {
        name      string
        setupMock func(repo *mocks.MockOrderRepo)
        req       *CreateOrderReq
        wantErr   error
    }{
        {
            name: "happy path",
            setupMock: func(repo *mocks.MockOrderRepo) {
                repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&Order{ID: 1}, nil)
            },
            req:     &CreateOrderReq{Amount: decimal.NewFromInt(100)},
            wantErr: nil,
        },
        // 更多 case...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            repo := mocks.NewMockOrderRepo(ctrl)
            tt.setupMock(repo)

            uc := NewOrderUsecase(repo)
            _, err := uc.Create(context.Background(), tt.req)

            if tt.wantErr != nil {
                require.Error(t, err)
                assert.True(t, errors.Is(err, tt.wantErr))
                return
            }
            require.NoError(t, err)
        })
    }
}
```

### 3. 自查清单（提交前勾完）

- [ ] 所有 case 都在同一个表里
- [ ] mock 是 mockgen 生成的（有 `//go:generate` 注释）
- [ ] 没有真 IO（无 http.Get / sql.Open / kafka 真实连接）
- [ ] 错误断言用 `errors.Is` / `errors.As`
- [ ] 时间/随机已注入
- [ ] 金额用 `decimal.Equal`
- [ ] 命名 `Test{对象}_{方法}_{场景}_{期望}`
- [ ] `go test -race ./...` 通过
- [ ] `go test -cover` 达标

## 可选模块（需在 App 里勾选安装）

| 模块 | 说明 |
|---|---|
| `references/test-structure.md` | 测试文件放哪 + 包名规则 |
| `references/table-driven.md` | 表驱动完整套路（含并行、子测试） |
| `references/mock-gomock.md` | mockgen + gomock EXPECT/Return/Times/InOrder 详解 |
| `references/assertions.md` | testify assert vs require 决策树 |
| `references/anti-patterns.md` | 常见反模式（time/rand/全局态/sleep 等） |
| `references/coverage.md` | 覆盖率命令 + exclude 规则 |
| `references/fixtures.md` | testdata/ + golden file + -update flag |
| `references/integration.md` | 集成测试 + build tag + dockertest |
| `assets/skeleton-usecase.go` | usecase 层测试骨架 |
| `assets/skeleton-http-handler.go` | HTTP handler 测试骨架（httptest） |
| `assets/skeleton-repo-sqlmock.go` | repo 层测试骨架（sqlmock） |

**默认全部不装**。在 Team Standards App → 「🧪 Go Unit Test Skill」卡片里按需勾选。
