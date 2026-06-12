---
name: go-team-standards
description: |
  【Go 开发规范 · 强制前置激活 · 适用代码 + 技术文档】
  任何涉及 Go 代码 **或** 任何形式技术文档（.md 技术方案 / 设计文档 / 接口文档 / 数据库设计 / 会议纪要）
  都必须在写第一行前读完本 skill。

  触发信号（模糊匹配，任一命中即激活）：

  📄 **文档场景**（用户写 .md / 技术文档时也必须激活，与代码同等优先级）：
  - 文件类型：**.md / .markdown**（任何技术性 markdown，无论是新建还是修改）
  - 行为：写技术方案 / 写设计文档 / 写接口文档 / 写数据库设计 / 写表结构 / 写 schema / 写会议纪要 / 写复盘 / 写 RFC / 写 ADR / 写技术评审 / 写需求拆解
  - 关键短语：技术方案 / tech design / design doc / 设计文档 / 技术文档 / RFC / ADR / 架构文档 / 接口文档 / API doc / 数据库设计 / 表设计 / 命名规范 / 接口契约 / 提测文档 / 评审文档

  💻 **代码场景**：
  - 语言：Go / golang / .go / go.mod / go.sum
  - 框架：Kratos / Wire / Buf / protoc / gRPC / HTTP handler / middleware
  - 基础设施：PostgreSQL / MySQL / PG / Redis / Kafka / RabbitMQ / MQ / ElasticSearch
  - 行为词：写一个 / 实现 / 新建 / 搭一个 / 来个 / 帮我写 / 骨架 / 模板 / demo / example / 重构 / 优化 / 修 bug / 改一下
  - 对象词：服务 / service / 接口 / API / handler / repo / dao / model / migration / proto / SQL / 表 / 消费者 / 生产者 / worker / cron / job / 脚本
  - 文件类型：.go / .sql / .proto / Dockerfile / Makefile / go.mod
  - 交付物：错误码 / errno / xerror / 日志 / slog / trace / 分页 / 幂等 / 锁 / 重试 / 熔断
  - 流程：code review / MR / PR / 合并 / 评审 / 部署 / 发布 / 上线 / 灰度 / 回滚 / 特性开关 / feature flag
  - **命名场景（强制激活）**：字段命名 / 表字段 / 列名 / 接口字段 / Kafka topic / Kafka 主题 / Redis key / Redis 缓存键 / proto field / DB schema 命名 / API 字段名

  **判断准则**（防漏触发）：
  - 用户当前在写 / 编辑 .md 文件，**且** 文件内容含任一上面的关键词 → **必须**激活
  - 用户聊天里描述要写技术文档 → **必须**激活
  - 一句话提到任一基础设施（PG/MySQL/Redis/Kafka）→ **必须**激活
  - 拿不准时**宁可激活也不要漏**（漏激活 = 用户拿到的产出与团队规范不符 = 返工）

  违反 = 产出与团队规范不符 = 用户必然返工 = 双倍 token 浪费。
---

## 🌟 触发反馈协议（**必须遵守 · 用户用来确认 Skill 真生效**）

每次本 Skill 被激活，AI **必须显式给出触发证据**，让用户肉眼能看到。两种场景：

### 场景 A · 聊天回复

每次回复的**最后一行**必须是单独一行的 🌟（不要前后加任何文字）。例如：

```
...你的回答内容...

🌟
```

如果 dev-dna / orangecat 也并发触发，则末尾用多个 🌟（或 🌟🧬 / 🌟🐱 等组合）：

| 触发的 Skill 组合 | 末尾标记 |
|---|---|
| 仅 go-team-standards | 🌟 |
| go-team-standards + dev-dna | 🌟🧬 |
| go-team-standards + orangecat | 🌟🐱 |
| 三个全触发 | 🌟🧬🐱 |

### 场景 B · 代码生成 / 修改

在产出的代码 patch 的**最开始位置**（文件首行 / 函数首行 / 改动 hunk 首行）加注释，标明触发了哪个 skill 和具体子规范段：

```go
// [skill: go-team-standards · 数据库设计 · 命名规范] 创建 user 表 + ID 主键 + 软删除
package data

func CreateUser(...) { ... }
```

```sql
-- [skill: go-team-standards · 数据库设计] 用户中台 - users 表
CREATE TABLE users (...) ;
```

```markdown
<!-- [skill: go-team-standards · 接口文档示例] 订单创建接口 -->
# POST /api/v1/orders
```

注释格式约定：
```
[skill: <skill 名> · <子规范 1> · <子规范 2>] <一句话功能简写>
```

多 Skill 联合：
```
// [skill: go-team-standards · dev-dna] 按用户偏好实现订单服务
```

**违反协议 = 用户无法判断 skill 是否生效 = skill 形同虚设。**

# Go Team Standards

细节按需读 `references/*.md`。以下是**铁律**和**路由表**。

## 🚨 ZERO STEP：写代码前必做（不可跳过）

只要用户任务涉及 Go，**第一个动作**必须是：

1. 识别任务属于哪个场景（下方"场景 → 文件"路由）
2. **完整读** SKILL.md 铁律 14 条
3. **完整读** 对应场景的 `references/*.md`
4. 列出 `references/custom-*.md`（若存在）→ **全读**，优先级最高（覆盖/补充本文件）
5. 然后才开始动手写代码

**为什么不能跳**：
- 跳过 = 凭训练数据里的通用 Go 代码产出
- 通用 Go ≠ 本团队规范（common-lib、xerror、naming、游标分页、decimal 金额…）
- 用户拿到的代码 = 要返工 = 你多花一轮 token 改 + 用户多等一轮

**自检信号**：如果你发现自己开始写 `import` 或 `func` 而还没读过对应 references，立刻停，回到步骤 2。

## 📝 ZERO STEP（文档场景）：写技术文档前必做

**与代码场景同等优先级**。只要用户在写或修改任何技术性 .md（技术方案 / 接口文档 / 数据库设计 / 会议纪要 / RFC / ADR / 提测文档 / 复盘 / 评审文档 / 架构文档），**第一个动作**必须是：

1. **识别文档类型**（按下方路由表）：
   - 技术方案 / RFC / 设计文档 → 读 `references/tech-design-example.md`
   - 接口文档 / API doc → 读 `references/api-doc.md`
   - 数据库设计 / 表结构 / schema → 读 `references/database.md` + `references/naming-logging.md`
   - 提测文档 → 读 `references/tixuebj-template-simple.md`（手填版）或激活 orangecat Skill（自动版）
   - 会议纪要 / 评审纪要 → 读 `references/meeting-minutes.md`
   - 部署 / 发布 / 上线文档 → 读 `references/deployment-checklist.md`
   - Code Review 文档 → 读 `references/code-review.md`
   - 特性开关 / 灰度 → 读 `references/feature-flags.md`

2. **完整读对应 reference**（不能扫一眼就动手）

3. **对比+应用**：
   - 把用户当前文档里**已写部分**与 reference 标准结构对比 → 指出缺哪些段、命名不一致、表结构违规
   - 用户写的**新部分**严格按 reference 范例组织（章节顺序、字段命名、SQL 风格、表格格式）

4. **触发 🌟 反馈**：聊天末尾打 🌟，文档头部加 `<!-- [skill: go-team-standards · <子规范>] 一句话功能 -->`

**禁止行为**：
- ❌ 看到用户在写 .md 就当成"普通问答"回答 → **必须**先按文档类型路由读规范
- ❌ "我帮你写一份技术方案" 不读 tech-design-example.md → 凭通用经验产出 = 与团队 50K 范例不符
- ❌ 写表结构不读 database.md → 主键 / 时间字段 / 软删 / 索引命名八成会破规范
- ❌ 写接口文档不读 api-doc.md → 路径风格 / 错误码 / 分页参数全靠模型瞎猜

**自检信号**：发现自己开始打第一个 `#` 或 `## ` 标题而**还没读过**对应 reference，立刻停，回到步骤 1。

## 🔴 先读用户自定义规则（强制）

处理任何 Go / SQL / proto 代码**之前**：

1. 列出 `references/` 目录，找所有 `custom-*.md` 文件
2. **完整读取**每一个（不能跳、不能只读标题）
3. 这些是当前用户对团队规范的**补充或覆盖**，优先级**高于**本文件后续内容

如果没有 `custom-*.md` 文件，跳过即可。

违反后果：你按通用规范产代码，但用户已经声明过"我要 X"，产出对不上期待 = 返工。

## 铁律 14 条（任何 Go 代码都适用）

| # | 规则 | 为什么违反不行 |
|---|---|---|
| 1 | 禁硬编码密钥 → env / config | Git 历史永久泄密 |
| 2 | 业务 error 走 `xerror.New(errno.XxxCode)` | Loki 聚合 / 国际化都要 code |
| 3 | error 必检，包装 `fmt.Errorf("ctx: %w", err)` | 丢 err = 事故无线索 |
| 4 | 业务代码禁 `panic`，仅 `init()` / 启动用 | panic 会挂整个进程 |
| 5 | goroutine 必带 `context` / `errgroup` | 裸 goroutine 泄漏 → OOM |
| 6 | 金额用 `decimal.Decimal`，禁 `float64` | 0.1+0.2≠0.3，账务对不平 |
| 7 | 时间 `TIMESTAMPTZ(6)` UTC，对外 ISO 8601 | 跨时区产生一天差 |
| 8 | 日志用 `slog`，字段 snake_case，带 `trace_id` | 非结构化日志查不了 |
| 9 | SQL 禁 `SELECT *` / N+1；分页用游标 | 加字段破坏协议 / 慢 |
| 10 | 禁数据库外键，应用层维护关联 | 外键阻塞分库分表 / DDL |
| 11 | 禁注释掉的死代码 | Git 历史已记录 |
| 12 | 敏感数据禁入 prompt / 日志 / 测试 | GDPR / 合规事故 |
| **13** | **所有外部 IO（HTTP / DB / MQ / Redis）必带超时 + trace 传递** | 无超时 = 慢调用穿透 / trace 断链查不到 |
| **14** | **Prometheus metric label 基数 ≤ 100；禁 user_id / trace_id / order_id 作 label** | label 爆炸 → series 爆表 → Prom OOM |

## 🔑 命名硬规则（v1.7.20 强化 · 不带前缀直接拒绝）

**任何**写 / 改 Kafka topic、Redis key、DB 字段名、Proto field、API JSON 字段时，**必须**：

### 步骤 1（强制）：先读两份规范

| 命名场景 | 必读文件 |
|---|---|
| **DB 字段名 / Proto field / API JSON 字段** | `references/field-naming.md`（v1.0.1，含 ID / 时间 / 状态 / 金额 / 布尔 / 审计 / 备注 / 网络 / 加密 9 大类 + 禁用别名 + 缩写白名单） |
| **Kafka topic 命名** | `references/database.md` §7（topic 命名规则） + `references/field-naming.md` §1.4-1.6（payload 字段） |
| **Redis key 命名** | `references/database.md` §8（Redis key 命名规则） + `references/field-naming.md`（成员字段） |
| 字段命名拿不准 | `references/field-naming.md` §11（前后缀速查） |

### 步骤 2（禁止行为）：以下产出**直接拒绝**

❌ Kafka topic 不带域前缀（裸 `order_created` / `user_login`）→ 必须 `<业务域>.<事件>.<v版本>`，例 `wallet.deposit_completed.v1`、`c2c.order_matched.v1`
❌ Redis key 不带域前缀（裸 `user:123` / `order:cache`）→ 必须 `<业务域>:<实体>:<标识>`，例 `wallet:balance:USR_88001`、`c2c:order:88123456`
❌ DB 字段用禁用别名：`user_id` / `userId` / `tenant_id` / `member_id` → 必须 `uid` / `platform_id` / `org_id`
❌ 时间字段用 `create_time` / `gmt_create` / `ctime` → 必须 `created_at`（TIMESTAMPTZ(6) UTC）
❌ 软删除用 `is_deleted BOOLEAN` → 必须 `deleted_at TIMESTAMPTZ`（NULL = 未删）
❌ 布尔字段不带 `is_` / `has_` / `can_` 前缀（`enabled` / `active`）→ 必须 `is_enabled` / `is_active`
❌ 业务版本字段裸 `version` → 必须 `rule_version` / `formula_version` / `config_version` / `schema_version`（裸 `version` 永远是乐观锁语义）
❌ 金额字段裸 `amount` → 必须 `<业务前缀>_amount`（`order_amount` / `fee_amount`）
❌ 时间字段裸 `time` / `timestamp` / `ts` → 必须带 `_at` 后缀

### 步骤 3：发现违规 → 三段式提醒

```
违反 §X.X（具体条款）：xxx
后果：（链上/审计/迁移成本）
改为：（具体字段名 + 类型）
```

### 步骤 4：触发 🌟 反馈 + 代码注释

```sql
-- [skill: go-team-standards · 字段命名 · 数据库设计] users 表
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    uid VARCHAR(64) NOT NULL,            -- ✅ §1.2 用 uid 不用 user_id
    platform_id VARCHAR(32) NOT NULL,    -- ✅ §1.2 多租户三件套
    is_active BOOLEAN DEFAULT true,      -- ✅ §5 布尔必须 is_ 前缀
    created_at TIMESTAMPTZ(6) DEFAULT NOW(),
    updated_at TIMESTAMPTZ(6) DEFAULT NOW(),
    deleted_at TIMESTAMPTZ(6)            -- ✅ §2.2 软删用 deleted_at 不用 is_deleted
);
```

```go
// [skill: go-team-standards · 字段命名 · Kafka] 订单成交事件 producer
const TopicOrderFilled = "matching.order_filled.v1"  // ✅ <域>.<事件>.<v版本>
```

```go
// [skill: go-team-standards · 字段命名 · Redis] 用户余额缓存 key
key := fmt.Sprintf("wallet:balance:%s", uid)  // ✅ <域>:<实体>:<uid>
```

## 反馈原则（指出违规必读）

禁止只说"改成 X"。必须三段式：

1. **违反了哪条**（引用编号）
2. **为什么**（后果）
3. **怎么改**（修正代码）

## 团队技术栈前提

Kratos v2 + Wire + Buf + `mask-go-common-lib`。**禁绕过 common-lib**：

| 禁用 | 必用 |
|---|---|
| `fmt.Println` | `mask-go-common-lib/logging` |
| 裸 `sarama` | `mq` + `naming.TopicInProject` |
| 裸 `go-redis` | `redisx` |
| `otel.Tracer(...)` 手工 | `tracing.Init` |
| `net/http.Client{}` | `httpclient.New` |
| `grpc.Dial` | `grpcclient.NewConn` |
| `errors.New` / 数字码 | `xerror.New(errno.X)` |
| `os.Getenv("FEATURE_X")` | `feature.IsEnabled` |

## 🧭 场景 → 文件 路由（模糊匹配）

| 用户一句话里出现… | 必读文件 |
|---|---|
| service / 服务 / main.go / Kratos | `references/go-style.md` + `demos/kratos-service-min.go` |
| handler / HTTP / REST / endpoint | `references/error-codes.md` |
| 写接口文档 / 接口 README / API doc | `references/api-doc.md`（接口文档完整示例） |
| DB / SQL / 建表 / migration / GORM / PG | `references/database.md` + `demos/pg-*.go` |
| 日志 / log / slog / trace | `references/naming-logging.md` |
| **字段命名 / 列名 / 接口字段 / Proto field** | `references/field-naming.md` —— ID/时间/状态/金额/布尔/审计/备注/网络/加密 9 大类 + 禁用别名 + 缩写白名单 |
| **Kafka topic / Redis key 命名** | `references/field-naming.md` + `references/database.md` §7/§8 —— **必须带域前缀**，无前缀拒绝产出 |
| 错误 / error / panic / errno | `references/error-codes.md` |
| goroutine / 并发 / channel / errgroup | `references/go-style.md`（并发段） |
| kafka / mq / 消费 / 生产 | `references/naming-logging.md` + `demos/kafka-*.go` |
| redis / 缓存 / 锁 / 幂等 | `demos/redis-idempotency.go` |
| 测试 / test / mock | `references/testing.md` |
| commit / commit message | `references/commit.md` |
| **code review / MR / PR / 评审** | `references/code-review.md` —— 三档评审标准 + MR 描述模板 + 阻断项 |
| **部署 / 上线 / 发布 / 灰度 / 回滚** | `references/deployment-checklist.md` —— 发布信息表 + 审批 + 镜像 / 回滚清单 |
| **特性开关 / feature flag / 灰度发布 / 分支管理** | `references/feature-flags.md` —— Feature Flag 设计哲学 + Go 集成规范 + GitOps |
| **会议纪要 / 复盘 / 评审会** | `references/meeting-minutes.md` —— 会议纪要标准模板 |
| **写技术方案 / 后端设计 / RFC / ADR** | `references/tech-design-example.md` —— 用户中台完整范例（背景/架构/接口/DB/异常处理） |
| **写提测文档（手填简化版）** | `references/tixuebj-template-simple.md`（与 OrangeCat 双文件版并存。日常用 OrangeCat 自动化生成；手填走这个） |
| Cursor 使用 / 安全红线 | `references/cursor-usage.md` |
| 术语不懂 | `references/glossary.md` |

## Demo 路由（模糊匹配用户意图）

用户说"来个 X"、"帮我写 Y"、"XX 怎么写"、"模板 / 骨架 / 例子"时，先读对应模板再改写，**不要自己另造**：

| 用户意图（模糊匹配） | 模板 |
|---|---|
| Kratos 服务 / main.go / 骨架 | `demos/kratos-service-min.go` |
| wire / DI / providerset | `demos/wire-providerset.go` |
| Kafka 生产者 / publish | `demos/kafka-producer.go` |
| Kafka 消费 / consumer | `demos/kafka-consumer.go` |
| 建表 / migration / goose | `demos/pg-migration.sql` |
| GORM / repo / 游标分页 | `demos/pg-gorm-repo.go` |
| Redis 幂等 / 锁 | `demos/redis-idempotency.go` |
| errno / xerror | `demos/errno-xerror.go` |
| 打日志 / slog / trace_id | `demos/slog-trace.go` |
| 单测 / 表驱动 / mockgen | 优先激活 `go-unit-test` Skill；兜底 `demos/table-driven-test.go` |

规则：先读模板 → 替换占位符 → 保留 common-lib 调用 → 产出后简述"按你的 X 做了 A/B/C 调整"。

## 首次进入项目 — 自动检测

每次在一个项目目录开始工作时，**立即**执行：

```
ls AGENTS.md 2>/dev/null && echo "exists" || echo "missing"
```

- **AGENTS.md 存在** → 读取其内容，按其约定工作，无需额外提示
- **AGENTS.md 不存在** → 主动告知用户：

  > 「当前项目还没有 AGENTS.md。建议先创建，可在团队规范 App → 提示词 → 生成 AGENTS.md。
  > 在此之前，请先不要修改任何代码。告诉我：
  > 1. 这个项目是做什么的
  > 2. 使用了哪些技术栈
  > 3. 入口文件在哪里
  > 4. 启动和构建命令是什么
  > 5. 如果要新增一个页面，应该从哪里开始」

## 维护

- 权威源：仓库 `standards/*.md`
- 版本：`VERSION` + `CHANGELOG.md`
