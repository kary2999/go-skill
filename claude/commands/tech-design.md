---
description: 写技术方案（按 tech-design-example.md 7 段范例）
argument-hint: <主题，如 "用户中台 v2 改造">
---

# 写技术方案：$ARGUMENTS

请按下列流程为我产出**技术方案**：

## 步骤

1. **必读**：`~/.claude/skills/go-team-standards/references/tech-design-example.md`（50K 完整范例）
2. **必读**：`~/.claude/skills/go-team-standards/references/database.md`（如果方案涉及数据库）
3. **必读**：`~/.claude/skills/go-team-standards/references/api-design.md`（如果方案涉及接口）
4. 按范例的 **7 段结构**输出：
   - 背景与目标
   - 现状分析（问题、痛点）
   - 总体方案（架构图、关键模块）
   - 数据库设计（如有）
   - 接口设计（如有）
   - 异常处理 / 兜底策略
   - 上线计划与回滚方案
   - 风险与待定事项

## 命名约束（强制）

- 用户主体字段一律 `uid`，禁 `user_id`
- 软删除一律 `deleted_at TIMESTAMPTZ`，禁 `is_deleted` BOOLEAN
- 时间字段一律 `_at` 后缀 + `TIMESTAMPTZ(6)` UTC
- 金额必带业务前缀（`order_amount`/`fee_amount`），禁裸 `amount`
- 业务版本字段必须前缀（`rule_version`/`config_version`）

## 输出要求

- 文档第一行下加注释：`<!-- [skill: go-team-standards · 技术方案] $ARGUMENTS -->`
- 末尾单独一行：🌟
