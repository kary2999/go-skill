---
title: "code-review Skill 使用说明"
version: "2.0.0"
audience: "团队所有人（开发 / Reviewer）"
last_modified: "2026-05-01"
---

# code-review Skill 使用说明

> 这是给**人**看的文档。AI 加载的是同目录 `SKILL.md`。

---

## 1. 这个 Skill 是干嘛的

**一句话**：AI 写完代码、或你粘代码进对话框时，**自动按团队 14 铁律 + 命名规范 review 一遍**，违规分级标出来 + 给修复方案。

**省你**：

- 不用每次都敲 `/review` —— AI 自己会走
- 不用记 14 铁律 —— skill 替你扫
- 不用怕 AI 写出违规代码偷偷塞给你 —— 它会自检

**和 `/review` 区别**：

| | code-review skill | `/review` 命令 |
|---|---|---|
| 触发 | **自动**，AI 自己判断该不该 review | 你主动敲 `/review` |
| 范围 | 当前代码片段 / AI 自产物 | git diff 或指定文件 |
| 输出 | 默认静默修复，违规多才出报告 | 总输出完整报告 |
| 用途 | 日常防违规 | PR 提交前 / 上线前自检 |

---

## 2. 怎么知道它生效了

### ✅ 看 emoji 标记

每次 skill 激活，**回复末尾会单独一行打**：

- `🔍` —— 只激活了 code-review
- `🌟🔍` —— code-review + go-team-standards 联合（最常见）
- `🌟🧬🔍` —— + dev-dna（个人偏好也激活了）

**没有 🔍** = skill 没触发或模型偷懒。

### ✅ 看代码 patch 头部

AI 改完代码后，diff 头会带：

```go
// [skill: code-review v2] 已自检 · P0 修 2 条 / P1 修 3 条 / 通过 8 项
```

### ✅ 看 TODO 注释

如果 AI 选了"模式 C"（不直接改，加注释让你决定），违规行上面会有：

```go
// TODO[review:P1·铁律#6 conf=85]: float64 → decimal.Decimal（金额必须 decimal）
balance float64
```

格式固定：`TODO[review:<级别>·<规则编号> conf=<分>]: <原> → <推荐>（<理由>）`

---

## 3. 触发场景速查

### 🟢 这些情况 AI 会**自动 review**

| 场景 | AI 行为 |
|---|---|
| AI 自己刚写完一段 .go / .sql / .proto / Dockerfile | 内心自检，违规直接修 |
| 你粘 ≥10 行代码 + 问"看下 / 这样行吗 / 帮我改改" | 输出完整 P0/P1/P2 报告 |
| 你说"review 这段 / 审一下 / 检查 / code review" | 同上，完整报告 |
| 你说"PR 提交前看看 / 上线前自检 / 合入前检查" | 同上 |

### 🔴 这些情况 AI **不会触发** review

| 场景 | 为什么不触发 |
|---|---|
| 你问"Go 切片底层怎么实现" | 这是教学问，不是 review 请求 |
| 你粘 ≤5 行伪代码示意 | 太短，没意义 review |
| 你在写 .md 技术文档 | 走 go-team-standards 即可 |
| 你说"先不 review" / "别评了" | 显式 opt-out |
| 同一段代码刚 review 过你没改 | 防重复打扰 |

### 🟡 拿不准时

AI 倾向**不触发**——错过比噪音好。你随时能敲 `/review` 主动叫。

---

## 4. 输出长啥样（实际示例）

### 示例 1 · 模式 A · 静默修复（最常见）

**你**：写一个 wallet 余额查询接口

**AI**（直接给你修好的代码）：
```go
// [skill: code-review v2] 已自检 · P0 修 1 条 / P1 修 2 条 / 通过 6 项

func GetBalance(ctx context.Context, uid string) (decimal.Decimal, error) {
    var balance decimal.Decimal
    if err := db.QueryRow(ctx, "SELECT balance FROM wallets WHERE uid = $1", uid).Scan(&balance); err != nil {
        return decimal.Zero, xerror.Wrap(err, errno.WalletQueryFailed)
    }
    return balance, nil
}
```

末尾：`🌟🔍`

**注意**：AI 没有先吐 `float64 balance` 的烂代码再让你骂——它内心已经走完 review 把 P0 修了。

---

### 示例 2 · 模式 B · 完整报告（你粘代码问"看下"）

**你**（粘了 30 行旧代码）：看下这段有啥问题

**AI**：

```markdown
## 团队规范 review

**总览**：5 处违规（P0 1 · P1 3 · P2 1）+ 2 条建议
**范围**：内联代码 · 30 行

### 🔴 P0 阻断（必改）

#### #1 硬编码密钥（铁律 #1，conf=98）
- **位置**：`第 5 行`
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
（同上格式...）

#### #3 errors.New 不走 xerror（铁律 #2，conf=92）
...

#### #4 SELECT * 全列查询（铁律 #9，conf=88）
...

### 🔵 P2 改进

#### #5 函数 124 行超长（conf=78）
...

<details>
<summary>💡 P3 建议（点开看）· 2 条</summary>

- 第 18 行变量 `tmp` 建议改 `pendingOrders`
- 第 25 行可用 `errors.Is` 替代字符串比对
</details>

### ✅ 自动确认通过

- 时间字段 _at + UTC（铁律 #7）
- 输出已带 trace_id（铁律 #13）

🌟🔍
```

---

### 示例 3 · 模式 C · TODO 注释（你说"标出来就行"）

**你**：这段是历史代码，先别改，把违规标出来我后面统一改

**AI**：

```go
// TODO[review:P0·铁律#1 conf=98]: 硬编码密钥 → os.Getenv("API_KEY")
const APIKey = "sk-live-abc123"

func transfer(amount float64) {
    // TODO[review:P0·铁律#6 conf=92]: float64 amount → decimal.Decimal（金额必须 decimal）
    // TODO[review:P1·命名§4.2 conf=80]: 裸 amount → order_amount / fee_amount
    
    // TODO[review:P1·铁律#13 conf=85]: http.Get 缺 ctx 和超时
    resp, _ := http.Get("https://api.example.com/")
    // TODO[review:P0·铁律#3 conf=95]: _ 丢弃 err
}
```

🌟🔍

---

## 5. 严重度怎么读

| 级 | 含义 | 例子 | 行动 |
|---|---|---|---|
| 🔴 **P0** | 必须改才能交付 | 密钥进 git · 金额 float · 业务 panic · SQL 注入 | **不改不让合**，AI 模式 A 会直接修掉 |
| 🟡 **P1** | 强烈建议改 | errors.New / 命名违规 / 无 ctx IO | 应改，但 deadline 紧可放宽 |
| 🔵 **P2** | 应该改但不阻断 | 日志缺 trace_id / 函数过长 | 视情况，下次重构一起 |
| 💡 **P3** | 可选优化 | 性能微优化 / 注释完善 | 折叠在 details 里，看看就行 |

---

## 6. 置信度怎么读（`conf=NN`）

每条违规挂 0-100 分。**默认只报**：

- P0 / P1 ≥ 70
- P2 ≥ 80
- P3 ≥ 90

**含义**：

| 分数段 | 含义 |
|---|---|
| 95-100 | 铁证（grep 字面量命中 / SQL 语法明显错） |
| 85-94 | 强证据（pattern + 上下文支持） |
| 70-84 | 比较确定（pattern 匹配 + 语义判断） |
| 50-69 | 怀疑，AI 不会报 |
| < 50 | 不会出现在输出里 |

**用途**：

- 看到 `conf=98` —— 不用想，改就完了
- 看到 `conf=72` —— 多看一眼场景，可能是 idiom 误判，可以反驳
- 看到 P2 而 `conf < 80` —— 不会出现（已被阈值过滤）

---

## 7. 怎么 opt-out（让 AI 别管这段）

### 7.1 单段代码豁免

在违规行**附近 5 行内**加：

```go
// review:ignore <理由：解释为什么这里合理>
const TestKey = "sk-test-abc"  // 测试用密钥，不是生产
```

skill 看到 `review:ignore` 注释会自动跳过该段，不报。

**注意**：必须给理由，光 `// review:ignore` 不写为啥不行——会被当成偷懒过滤。

### 7.2 当次对话不 review

直接说：
- "先不 review"
- "别评了，直接给我代码"
- "这段不用 review"

### 7.3 文件类型自动放宽

skill 自动对这些文件**降低严格度**：

- `*_test.go` —— 测试 fixture 命名可豁免（`user_id` 在测试里 OK）
- `testdata/` / `fixtures/` —— 测试数据
- `migrations/` —— 旧迁移脚本，旧字段保留为兼容
- `*.gen.go` / `*.pb.go` —— 生成代码不评

---

## 8. 主动调用 `/review`

如果你想**强制**让 AI review（不依赖自动触发）：

### 8.1 review git diff（PR 提交前最常用）

```
/review
```
或显式：
```
/review diff
```

→ AI 跑 `git diff origin/master...HEAD`，对所有改动走完整 review。

### 8.2 review 指定文件

```
/review internal/biz/wallet.go
```

→ AI 读这个文件，区分新增 vs 已存在，分别评。

### 8.3 review 内联代码

```
/review

func transfer(amt float64) { ... }
```

→ AI review 你给的这段。

---

## 9. 常见问题

### Q1：装完没反应？

```bash
# 1. 确认装上了
ls ~/.claude/skills/code-review/SKILL.md

# 2. 必做 —— Cmd+Q 彻底退出 Claude Code / Cursor 再开
#    Skill 只在启动时加载一次
```

### Q2：AI 写完代码不打 🔍？

可能原因：

- 模型偷懒（量化的 7B 小模型常见）—— 换大模型或敲 `/review` 强制
- 触发不命中（写的不是代码而是文档 / 配置）—— 正常
- skill 没装上 —— 跑 App 里「覆盖检查」

### Q3：报了一堆 P3 噪音？

不应该。P3 默认置信度阈值 90，应该很少触发。如果泛滥说明：

- 你的代码上下文 AI 不懂（建议补 CLAUDE.md 项目说明）
- 模型版本太老 —— 升级到 Claude Sonnet 4.6+ / Cursor latest

### Q4：把 idiom 当 bug 报了

应该不会，skill v2.0 §4 有"不报清单"防这个。如果还是报了：

1. 在违规行附近加 `// review:ignore <这是 sync.Once idiom>`
2. 反馈给团队规范 owner，下版本加进不报清单

### Q5：和 `/review` 重复了？

不重复：

- **skill** = 默认开的"日常体检"——AI 写完就扫一遍，免你操心
- **`/review`** = "正式体检"——PR/MR 提交前，你主动叫，输出完整报告归档

两个互补。skill 拦掉 80% 违规进不到代码里，`/review` 兜底关键节点。

### Q6：能不能改阈值（让它更严 / 更松）？

当前不支持运行时改阈值。如果要改，编辑 `~/.claude/skills/code-review/SKILL.md` 的 §3 数值，然后 Cmd+Q 重启。

---

## 10. 出问题反馈给谁

App 里「日志控制台」看 skill 加载状况；
规范本身有问题反馈给团队规范 owner（issue / IM）。

---

## 附：相关文件

| 文件 | 作用 |
|---|---|
| `~/.claude/skills/code-review/SKILL.md` | AI 加载的主 prompt（v2.0） |
| `~/.claude/skills/code-review/references/fix-examples.md` | 常见违规修复对照表 |
| `~/.claude/skills/code-review/USAGE.md` | **本文档**（给人看） |
| `~/.claude/commands/review.md` | `/review` 主动命令 |
| `~/.claude/skills/go-team-standards/SKILL.md` | 14 铁律本体 |
| `~/.claude/skills/go-team-standards/references/全局统一字段命名规范.md` | 命名规范本体 |
