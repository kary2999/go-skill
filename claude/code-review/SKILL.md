---
name: code-review
version: "2.0.0"
last_modified: "2026-05-01"
description: |
  【代码评审 · 自动激活，带置信度过滤 + 严重度分级 + 跳过清单】
  AI 自己写完代码 / 用户粘代码 → 自动按团队规范 review，违规分级 + 给修复方案。
  与 `/review` slash command 共享同一套规则，但 `/review` 由用户主动调，本 Skill 默认被动。

  ✅ 强触发（不等用户问）：
  - AI 刚生成 .go / .sql / .proto / .ts / Dockerfile / Makefile 代码 → 交付前自检
  - 用户粘 ≥10 行实际代码 + 模糊问句（"看下"/"这样行吗"/"帮我改"）

  ✅ 显式触发：
  - "review / 审一下 / 检查 / 看看 / code review"
  - "PR / MR / 评审 / 上线前自检 / 合入前检查"

  ❌ 不触发（避免噪音）：
  - 用户问语法/语言知识（"切片底层怎么实现"）—— 不是 review 场景
  - ≤5 行伪代码 / 示意片段
  - 用户在写 .md 文档（走 go-team-standards）
  - 用户已说"先不 review"
  - 同一段代码已 review 过且未变（防重复打扰）

  🤔 拿不准时倾向**不触发**——错过比噪音好；用户能随时敲 `/review` 主动调。
---

# Code Review Skill v2.0

## 0. 触发决策（最优先判定）

**不要看一眼就开干**。先按上面的 description 判断当前消息属于哪类：

| 信号 | 动作 |
|---|---|
| 强触发命中 | 走完整流程（§1→§5） |
| 显式触发命中 | 走完整流程，且**总是输出 review 报告**（模式 B） |
| 不触发命中 | 直接答用户问题，**不要硬塞 review** |
| 拿不准 | 不触发；末尾不打 🔍 |

> **常见过激活反模式**：用户粘了 3 行代码问"这是什么意思" → 这是教学问，不是 review 请求。直接讲解，不要扯铁律。

---

## 1. Pre-flight（review 前必做，不可跳）

按顺序：

1. **读规范**：`~/.claude/skills/go-team-standards/SKILL.md`（14 条铁律 + 路由）
2. **若涉及命名**（DB schema / proto / API JSON）：
   `~/.claude/skills/go-team-standards/references/全局统一字段命名规范.md`
3. **若 dev-dna 已激活**：参考 `~/.claude/skills/dev-dna/references/profile.md`
4. **判断范围**：
   - 用户给了 git diff / `/review diff` → **只评 diff 内容**，旧代码不动
   - 用户给了完整文件 → 评全文件，但**区分新增 vs 已存在**违规（已存在的标"待重构"，不算阻断）
   - AI 自己刚产出 → 评本次产出
5. **判断文件类型**（决定严格度）：
   - `*_test.go` / `testdata/` / `fixtures/` / `migrations/` → **放宽**（测试 fixture 命名可豁免，迁移脚本旧字段保留）
   - 业务代码 → 严格

**禁止**：没读规范就下结论 / 把 idiom 当 bug 报。

---

## 2. 严重度分级（每条违规打标）

| 级 | 含义 | 例子 |
|---|---|---|
| **P0 阻断** | 必须改才能交付 | 硬编码密钥 · SQL 注入 · 金额 float64 · 业务路径 panic · 敏感数据进日志 |
| **P1 重要** | 强烈建议改 | errors.New 不走 xerror · 命名规范违反 · 无 ctx 调 IO · SELECT * |
| **P2 改进** | 应该改但不阻断 | 日志缺 trace_id · 变量名不达意 · 函数过长（>80 行） |
| **P3 建议** | 可选优化 | 性能微优化 · 重构机会 · 注释完善 |

**默认输出 P0 + P1 + P2**；P3 折叠在 `<details>` 里（不抢主线）。

---

## 3. 置信度自评（0-100，过滤吹毛求疵）

每条违规挂置信度。计算依据：
- 100 = 铁证（grep 字面量命中 / SQL 语法明显错）
- 85 = 强证据（pattern 匹配 + 上下文支持）
- 70 = 比较确定（pattern 匹配但需要语义理解）
- 50 = 怀疑（看起来像，但可能是 idiom）
- 30 = 拿不准
- < 30 = 不要报

**默认阈值**：
- P0 / P1：confidence ≥ 70 才报
- P2：confidence ≥ 80 才报（更严，防止 P2 噪音淹没 P0）
- P3：confidence ≥ 90 才报

> 自查：每条违规问自己"如果用户回我'你确定？'，我能不能列出 3 条具体证据？"——列不出 = confidence 不够，撤回。

---

## 4. 不报清单（False Positive 黑名单）

**绝对不报**：
- 已存在代码不在本次改动范围内（指出"待重构"标注但不计违规）
- linter 能逮的（`gofmt` / `golangci-lint` / `goimports` 那批 —— 缩进、未用 import、空行）
- 不在团队规范里的命名争议（如 `Get*` vs `Fetch*` 这种纯口味）
- 看起来像 bug 但是 idiom：
  - `sync.Once` 的 `var once sync.Once` 不是死变量
  - `context.Background()` 在 main / 测试 / 启动期合法
  - `defer close(ch)` 在 producer goroutine 是正确模式
  - errgroup 的 `_ = g.Wait()` 在已经 select 处理 err 的场景合法
- 用户在代码里明确 opt-out：`// review:ignore <理由>` 上下相邻 5 行内
- 重复打扰：同一段代码上一轮已 review 过且没改 —— 不再重报，仅 1 行提醒
- 历史遗留代码（git blame > 6 个月且本次未改）—— 转为 P3 建议

---

## 5. 输出三模式（按场景自动选）

### 模式 A · 静默自检（默认）
**触发**：AI 自己刚产出代码、还没交付
- 内心走完 §1→§4
- 发现 P0/P1 → **直接修掉**重新产出，不要先吐烂代码等用户骂
- 末尾单独一行 `🔍`（与 go-team-standards 联合时 `🌟🔍`）
- 代码 patch 头部加：
  ```
  // [skill: code-review v2] 已自检 · P0 修 N 条 / P1 修 M 条 / 通过 K 项
  ```

### 模式 B · 完整 review 报告
**触发**：用户粘代码问"看下" / 显式叫 review / `/review`

```markdown
## 团队规范 review

**总览**：N 处违规（P0 a · P1 b · P2 c）+ d 条建议 · review 范围：<diff/file/snippet>

### 🔴 P0 阻断（必改）

#### #1 硬编码密钥（铁律 #1，conf=98）
- **位置**：`internal/config.go:23`
- **代码**：
  ```go
  const APIKey = "sk-live-abc123"
  ```
- **为什么不行**：源码进 git → 任何人 clone 都拿到密钥
- **怎么改**：
  ```go
  var APIKey = os.Getenv("API_KEY")
  ```

### 🟡 P1 重要（强烈建议）

#### #2 用户字段命名违反 §1.2（conf=85）
（同上格式）

### 🔵 P2 改进

#### #3 ...

<details>
<summary>💡 P3 建议（点开看）· 2 条</summary>

- ...
</details>

### ✅ 自动确认通过
- 错误处理走 xerror（铁律 #2）
- 时间字段 _at + UTC（铁律 #7）
- 输出已带 trace_id（铁律 #13）

### ⚠️ 已存在问题（本次未改，建议后续重构）
- `internal/old.go:45` 用了 SELECT * —— 历史代码，不阻断本次

末尾：`🔍`（联合则 `🌟🔍`）
```

### 模式 C · TODO 注释模式
**触发**：重写成本高 / 用户已粘大段历史代码 / 用户说"标出来就行别动"

在违规行**正上方**加：
```go
// TODO[review:P1·铁律#6 conf=85]: float64 → decimal.Decimal（金额必须 decimal）
balance float64
```

格式硬约束：
```
TODO[review:<级别>·<规则编号> conf=<分>]: <原> → <推荐>（<一句话理由>）
```

---

## 6. 14 条铁律 + 命名规范快查表

> 详细修复对照见 `references/fix-examples.md`。

| # | 铁律 | grep / 检测要点 |
|---|---|---|
| 1 | 禁硬编码密钥 | `sk-` / `kgb_` / `password=` / `Bearer ` 字面量 |
| 2 | error 走 xerror | `errors.New(` 直接调（除非 init 阶段） |
| 3 | error 必检 + `%w` | `_ = err` / 跨层 `fmt.Errorf` 不带 `%w` |
| 4 | 业务禁 panic | `panic(` 不在 main/init/启动路径 |
| 5 | goroutine 必带 ctx | `go func()` 无 ctx 形参 |
| 6 | 金额 decimal | `float64` + 字段名含 amount/balance/fee/price |
| 7 | 时间 _at + UTC | `time.Now()` 直调 / 字段非 `_at` 后缀 |
| 8 | 日志 slog + trace | `fmt.Println` / `log.Printf` 当业务日志 |
| 9 | 禁 `SELECT *` / N+1 | `SELECT \*` / 循环内 `db.Query` / OFFSET 分页 |
| 10 | 禁 DB 外键 | DDL `FOREIGN KEY` / GORM `foreignKey` tag |
| 11 | 禁死代码 | 大段 `//` 注释掉的代码 |
| 12 | 敏感数据禁入日志 | log/test fixture 含 phone/email/id_card |
| 13 | 外部 IO 必带超时 + ctx | `http.Get(` / `db.Exec(` 不传 ctx |
| 14 | metric label 基数 ≤ 100 | label 含 user_id/order_id/trace_id |

| 命名违规 | 推荐 | 规范段 |
|---|---|---|
| `user_id` / `userId` / `member_id` | `uid` | §1.2 |
| `gmt_create` / `create_time` / `ctime` | `created_at` | §2.2 |
| `gmt_modified` / `update_time` | `updated_at` | §2.2 |
| `is_deleted` BOOLEAN | `deleted_at TIMESTAMPTZ` | §5.2 |
| 裸 `amount` | `<业务>_amount`（`order_amount`） | §4.2 |
| `vol` / `size` / `amt` | `quantity` / `_qty` | §4.3 |
| 裸 `version`（业务） | `rule_version` / `config_version` | §6.3 |
| 布尔非 `is_` 前缀 | `is_<x>_enabled` | §5.1 |
| 裸 `time` / `ts` 业务列 | `<动作>_at` | §2.1 |
| `txid` / `transaction_hash` | `tx_hash` | §1.5 |
| `ip_addr` / `login_ip` | `client_ip` | §8 |
| 裸 `request_id` 跨域 | `<域>_request_id` | §1.6 |

---

## 7. 语气准则（学 wshobson 的 constructive 路线）

- **指出问题不指责人**：写"这里 float64 会丢精度"，**不**写"你又用 float 了"
- **每条违规给修复方案**，不只说"不行"
- **承认权衡**：性能/可读性冲突时说"两边都对，团队规范选 X 是因为 Y"
- **平衡严格 vs 推进**：deadline 紧时降级 P1→P2 也行，但 P0 不退让

---

## 8. 与其他 Skill 的边界

| Skill | 何时调 | 区别 |
|---|---|---|
| **code-review**（本，被动 v2.0） | AI 写完 / 用户贴代码 | 默认开，自动跑 |
| `/review`（主动 command） | 用户敲 `/review [diff/file]` | 用户控制范围 + 总走模式 B |
| `go-team-standards`（被动） | 写代码前 → 读铁律 | 写之前查；本 skill 是写之后查 |
| `dev-dna`（被动） | 用户偏好 | review 时也参考，但**铁律 > 偏好** |

---

## 9. 禁止行为（硬约束）

- ❌ AI 自己写完代码不自检就吐给用户（默认必走模式 A）
- ❌ 没读 SKILL.md 就下结论 / 凭印象 grep
- ❌ 把 idiom 当 bug 报（先查 §4 不报清单）
- ❌ 全部违规一锅炖、不分 P0-P3
- ❌ 不挂置信度 / 置信度低于阈值还硬报
- ❌ TODO 注释格式不统一（必须 `TODO[review:<级>·<规则> conf=<分>]: ... → ...`）
- ❌ review 完不打 🔍
- ❌ 在不触发场景（语法问 / 教学问 / 短伪代码）硬塞 review
