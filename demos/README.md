# Demos — 轻量最佳实践片段

这不是完整可跑项目，是**能直接复制到 `mask-go-sample-service` 风格脚手架的代码片段**。每个 demo 聚焦一个场景，优先调用 `mask-go-common-lib` 封装，**不绕开团队基础设施**。

## 何时用
- 用户在 Cursor / Claude Code 里说类似"帮我写个 kafka 生产者"、"来个 redis 幂等示例"、"pg 迁移模板"、"新建 kratos 服务"、"errno 怎么用"…… AI 应该识别意图并参考对应 demo 产出代码。**模糊匹配即可**，不要求关键词完全一致。

## Demo 索引

| 文件 | 场景 | 触发关键词（模糊） |
|---|---|---|
| `kratos-service-min.go` | 最小 Kratos HTTP+gRPC 服务骨架 | kratos / 新服务 / service 骨架 / 启动入口 |
| `wire-providerset.go` | Wire ProviderSet + 构造函数注入 | wire / DI / 依赖注入 / providerset |
| `kafka-producer.go` | 用 `mq.NewProducer` + 命名规范发消息 | kafka / 生产者 / 发消息 / producer / topic |
| `kafka-consumer.go` | 用 `mq.NewConsumer` 消费 + 优雅退出 | kafka / 消费者 / consumer / 订阅 |
| `pg-migration.sql` | goose up/down 模板，含必备字段 + 索引 + COMMENT | 建表 / 迁移 / migration / goose / schema |
| `pg-gorm-repo.go` | GORM repo 模式：ctx + timeout + 软删 + 游标分页 | gorm / 仓储 / repo / crud / 查询 / pg 读写 |
| `redis-idempotency.go` | `redisx` 幂等键 + 分布式锁（SETNX + lua 解锁） | redis / 幂等 / lock / 锁 / idempotency |
| `errno-xerror.go` | 错误码声明 + `xerror.New` 返回 + 包装 | errno / xerror / 错误码 / return error |
| `slog-trace.go` | `slog` + OTel ctx 提取 trace_id/span_id | 日志 / slog / trace_id / 打日志 |
| `table-driven-test.go` | 表驱动单测 + mockgen 骨架 | 单测 / 测试 / 表驱动 / mock / 覆盖率 |

## 使用约定

1. Demo 里的 import 路径按 `mask-go-common-lib` 真实包路径填写（`github.com/<org>/mask-go-common-lib/...`），若你团队 module path 不同，复制后整体替换。
2. 所有 demo 都假设 `context.Context` 已携带 OTel span、`trace_id`，上游 middleware 已注入。
3. 命名参数（`orderSvc`、`userID`）请按业务替换；类型和方法签名是模板的重点。
4. 所有 demo **不包含完整 main + docker-compose**。本地跑服务请参考 `mask-go-sample-service`。
