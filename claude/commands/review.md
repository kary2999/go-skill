---
description: 按团队 14 条铁律 + 命名规范 + dev-dna 偏好 review 改动 / 文件（带 P0-P3 分级 + 置信度）
argument-hint: <可选：文件路径 / "diff" / 留空>
---

# 团队规范 review：$ARGUMENTS

> 这是 `code-review` Skill v2.0 的**主动调用入口**。共享同一套规则，默认走"模式 B 完整 review 报告"。

## 1. 取代码

- $ARGUMENTS 为空 / `diff` → `git diff origin/master...HEAD`
- $ARGUMENTS 是文件路径 → 读该文件
- 否则视为内联代码

## 2. Pre-flight（必做）

按顺序读：
1. `~/.claude/skills/go-team-standards/SKILL.md`（14 铁律）
2. 含命名 → `~/.claude/skills/go-team-standards/references/全局统一字段命名规范.md`
3. dev-dna 已激活 → `~/.claude/skills/dev-dna/references/profile.md`（偏好，铁律冲突时**铁律优先**）

判断**范围**（diff vs 全文件）和**文件类型**（test/migration 放宽，业务严格）。

## 3. 执行（按 code-review skill v2.0 §1-§7 走）

为每条违规标：
- **严重度**：P0 阻断 / P1 重要 / P2 改进 / P3 建议
- **置信度**：0-100（默认阈值 P0/P1≥70 · P2≥80 · P3≥90）

不报清单（必须过滤）：
- 已存在代码不在本次 diff 内 → 标"待重构"不阻断
- linter 能逮的（gofmt / golangci）
- idiom（sync.Once / context.Background() in main / errgroup _ = g.Wait()）
- `// review:ignore` 注释相邻 5 行
- 不在团队规范里的纯口味命名

## 4. 输出格式

```markdown
## 团队规范 review

**总览**：N 处违规（P0 a · P1 b · P2 c）+ d 条建议
**范围**：<diff / 文件 / 内联> · <X 文件 · Y 行>

### 🔴 P0 阻断（必改）

#### #1 <规则编号 · conf=NN>
- **位置**：<file:line>
- **代码**：```<lang> ... ```
- **为什么不行**：<具体后果，不要"不规范"这种空话>
- **怎么改**：```<lang> ... ```

### 🟡 P1 重要（强烈建议）
（同上）

### 🔵 P2 改进
（同上）

<details>
<summary>💡 P3 建议（点开看）· N 条</summary>
- ...
</details>

### ✅ 自动确认通过
- ...

### ⚠️ 已存在问题（本次未改，建议后续）
- <file:line>: <问题> —— 历史代码，不阻断本次
```

末尾单独一行：`🌟🔍`（go-team-standards + code-review 联合）

## 5. 14 铁律 + 命名规范速查

详细见 `code-review` skill 的 §6 表格。常踩雷：

- **铁律 #1** 硬编码密钥（grep `sk-` / `kgb_` / `password=`）
- **铁律 #2** errors.New 不走 xerror
- **铁律 #6** 金额 float64
- **铁律 #9** SELECT * / N+1 / OFFSET 分页
- **铁律 #13** http.Get / db.Exec 无 ctx 无超时
- **命名 §1.2** user_id → uid
- **命名 §2.2** gmt_create → created_at
- **命名 §5.2** is_deleted BOOLEAN → deleted_at TIMESTAMPTZ
- **命名 §4.2** 裸 amount → 业务前缀（order_amount）

## 6. 禁止

- ❌ 没读规范就 review
- ❌ 不分 P0-P3 一锅炖
- ❌ 不挂 confidence 分数
- ❌ 把 idiom 当 bug 报
- ❌ review 完不打 🌟🔍
