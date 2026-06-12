# 单元测试 vs 集成测试

## 边界

| 特征 | 单元测试 | 集成测试 |
|---|---|---|
| 测什么 | 单个函数 / 单个结构体的逻辑 | 多个组件协作 / 真依赖 |
| 外部依赖 | **禁用真 DB/MQ/HTTP** | 允许真依赖（容器化） |
| 速度 | 毫秒级 | 秒级可接受 |
| 稳定性 | 必须稳定 | 允许偶发（重试） |
| 跑频率 | 每次 commit / 每次保存 | PR 合入前 / 每日 nightly |

## build tag 区分

```go
//go:build integration

package biz_test

import "testing"

func TestOrderIntegration_CreateAndFetch(t *testing.T) {
    // 真连 PG
}
```

跑法：
```bash
# 只跑单元
go test ./...

# 只跑集成
go test -tags=integration ./...

# 一起跑
go test -tags=integration ./...
```

## 命名约定

- 单元测试文件：`order_test.go`
- 集成测试文件：`order_integration_test.go`（加 build tag）
- 测试函数前缀：`TestXxxIntegration_`（便于 grep / -run 筛）

## dockertest 模板

```go
//go:build integration

package data_test

import (
    "testing"
    "github.com/ory/dockertest/v3"
)

func TestOrderRepo_Create(t *testing.T) {
    pool, err := dockertest.NewPool("")
    require.NoError(t, err)
    resource, err := pool.Run("postgres", "15", []string{
        "POSTGRES_PASSWORD=secret",
        "POSTGRES_DB=test",
    })
    require.NoError(t, err)
    t.Cleanup(func() { pool.Purge(resource) })

    dsn := fmt.Sprintf("postgres://postgres:secret@localhost:%s/test?sslmode=disable",
        resource.GetPort("5432/tcp"))

    var db *sql.DB
    require.NoError(t, pool.Retry(func() error {
        var e error
        db, e = sql.Open("postgres", dsn)
        if e != nil {
            return e
        }
        return db.Ping()
    }))

    // 跑 migration
    require.NoError(t, runMigrations(db))

    repo := NewOrderRepo(db)
    // ... 真实测试
}
```

## sqlmock（轻量替代）

不想起容器时：
```go
import "github.com/DATA-DOG/go-sqlmock"

db, mock, _ := sqlmock.New()
mock.ExpectQuery("SELECT .* FROM orders").WillReturnRows(
    sqlmock.NewRows([]string{"id"}).AddRow(1))
```

注意：sqlmock 验证的是 SQL 字符串 + 参数，**不是真 DB 行为**。如果你的 SQL 里有 PG 特定语法（JSONB / CTE），建议用 dockertest。

## CI 分轨

```yaml
# .gitlab-ci.yml
unit:
  script: go test -race -cover ./...

integration:
  script: go test -tags=integration ./...
  services:
    - postgres:15
    - redis:7
```

单元挂 = 阻断合并；集成挂 = 允许手动重跑（偶发容忍）。
