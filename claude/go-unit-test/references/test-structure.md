# 测试文件结构

## 文件放哪

| 被测 | 测试文件 |
|---|---|
| `internal/biz/order.go` | `internal/biz/order_test.go` |
| `internal/data/order_repo.go` | `internal/data/order_repo_test.go` |
| `internal/service/order.go` | `internal/service/order_test.go` |
| 工具函数 `pkg/util/time.go` | `pkg/util/time_test.go` |

与源文件**同目录同级**，不另开 `tests/` 顶级目录。

## 包名约定

| 场景 | 包名 | 原因 |
|---|---|---|
| 黑盒测试（只测公开 API） | `biz_test`（外部包） | 避免测试耦合私有实现 |
| 白盒测试（要测私有函数） | `biz`（同包） | 可访问私有标识符 |
| 大多数 usecase/service 层 | **优先黑盒** `xxx_test` | 强制面向接口设计 |

## helper 文件

共用的 setup 放：
- `{pkg}_testhelper_test.go` —— 包级 helper（`func newTestUsecase(t *testing.T) *Usecase`）
- `testdata/*.json` / `testdata/*.golden` —— 测试数据

## mock 文件放哪

```
internal/biz/
├── order.go
├── order_repo.go          ← 定义 OrderRepo 接口
├── order_test.go
└── mocks/
    └── mock_order_repo.go ← mockgen 生成，package mocks
```

**禁止放在 `internal/biz/` 本目录**（会污染被测包名）。

## 命名

| 对象 | 模式 | 例 |
|---|---|---|
| 顶层测试函数 | `Test{类型}_{方法}` | `TestOrderUsecase_Create` |
| 表驱动子 case | `{场景}_{期望结果}` | `"idempotent_replay_returns_existing"` |
| Benchmark | `Benchmark{类型}_{方法}` | `BenchmarkOrderUsecase_Create` |
| Example | `Example{类型}_{方法}` | `ExampleOrderUsecase_Create` |

子 case 名用 **小写 + 下划线**，便于 `go test -run 'TestOrderUsecase_Create/idempotent'` 精准跑。
