# Skill 有效性测试用例

这些文件是**故意写错**的 Go 代码，用来测试 Claude Code / Cursor 的 `go-team-standards` Skill 能否检出违规。

> ⚠️ **这些文件只读不运行**：
> - 每个文件头带 `//go:build skill_test_only` build tag
> - `go build` / `go test` 默认**不会编译它们**
> - 本目录（`.docs/`）**不进 git**（脚手架会自动加到 `.gitignore`）

---

## 怎么用

### 验证"主动触发"能力

在 Claude Code / Cursor 里打开任意一个文件，说一句**普通的**话：

> "这段代码帮我看看"

或：

> "review 一下"

**期望行为**：Skill 自动激活（无需你点名 `@go-team-standards`），按三段式指出违规：
1. 违反了哪条铁律（引用编号）
2. 为什么违反不行（后果）
3. 怎么改（修正代码）

### 验证"前置激活"能力

新开一个对话，直接说：

> "帮我实现一个订单服务，要支持幂等创建"

**期望行为**：在写第一行代码**之前**，模型就先读 SKILL.md + 对应 references。如果它上来就写代码而没提规范 —— 说明触发器太弱，需要加强。

---

## 用例清单（对照铁律 1-14 + 单测反模式）

| 文件 | 违反铁律 | 期望 AI 指出 |
|---|---|---|
| `01_hardcoded_secret.go` | #1 密钥 | 指出硬编码 API Key，建议走 config / env |
| `02_bare_errors_new.go` | #2 错误码 | 指出裸 `errors.New`，改 `xerror.New(errno.X)` |
| `03_silent_error.go` | #3 丢 error | 指出 `_ = err` 吞错，改为 `fmt.Errorf("ctx: %w", err)` 向上包装 |
| `04_business_panic.go` | #4 panic | 指出业务代码 panic 会挂整进程 |
| `05_naked_goroutine.go` | #5 goroutine | 指出无 ctx 的裸 goroutine 会泄漏 |
| `06_float_money.go` | #6 金额 | 指出 float64 存金额，改 decimal.Decimal |
| `07_local_time.go` | #7 时间 | 指出本地时间不带时区，改 UTC |
| `08_fmt_println_log.go` | #8 日志 | 指出 fmt.Println 打日志，改 slog + trace_id |
| `09_select_star_n_plus_1.go` | #9 SQL | 指出 SELECT * + N+1 + OFFSET 分页 |
| `10_fk_constraint.go` | #10 外键 | 指出不要 DB 外键，应用层维护关联 |
| `11_dead_commented_code.go` | #11 死代码 | 指出注释掉的代码应删除 |
| `12_log_sensitive_data.go` | #12 敏感数据 | 指出日志打印手机号 |
| `13_no_timeout_http.go` | #13 IO 超时 | 指出无超时 http.Get / 未传 ctx |
| `14_high_card_metric.go` | #14 label 基数 | 指出 user_id 作 label 会爆 series |
| `15_test_antipattern.go` | 单测反模式 | 指出手写 mock、time.Now() 直调、err.Error() 字符串断言 |

---

## 怎么判断 Skill 生效

| 情况 | 现象 | 处理 |
|---|---|---|
| ✅ 全通过 | 15 个用例 AI 都能按三段式指出违规 | Skill 工作正常 |
| ⚠️ 部分漏指 | 比如只指 6/15 | 更新 Skill（Team Standards App → 重新安装 go-team-standards），然后 Cmd+Q 彻底重启 AI 客户端 |
| ❌ 完全不触发 | AI 直接开始写代码 / 完全不提规范 | 检查 `~/.claude/skills/go-team-standards/SKILL.md` 是否存在；Cursor 侧看 `~/.cursor/rules/00-iron-laws.mdc` 是否存在 |
| 🔄 每次都漏指同一条 | 说明 description 没匹配到对应关键词 | 反馈给规范维护人（团队 channel） |

---

## 注意

- 这些代码**只用于测试规范 Skill**，不是学习资料（都是反面教材，抄了就破规）
- 删除本目录不影响项目构建
- 新同学入职后，可以对着这些文件跑一遍 AI，熟悉团队规范
