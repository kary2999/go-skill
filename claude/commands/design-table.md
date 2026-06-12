---
description: 设计 SQL 表（按 database.md + 全局命名规范）
argument-hint: <表名 + 业务说明，如 "orders 订单表，含币种、价格、状态">
---

# 设计表：$ARGUMENTS

请按团队规范为我产出**完整的 CREATE TABLE 语句**和说明。

## 步骤

1. **必读**：`~/.claude/skills/go-team-standards/references/database.md`
2. **必读**：`~/.claude/skills/go-team-standards/references/naming-logging.md`
3. **必读**：（如有）`~/.claude/skills/go-team-standards/references/全局统一字段命名规范.md`

## 强制规则

### 必备字段（每张表）

```sql
id          BIGSERIAL PRIMARY KEY,
uid         VARCHAR(64) NOT NULL,        -- 多租户：用户主体（不是 user_id）
platform_id VARCHAR(32) NOT NULL,        -- 多租户：平台标识
created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT NOW(),
updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT NOW(),
deleted_at  TIMESTAMPTZ(6),              -- 软删除（禁用 is_deleted BOOLEAN）
created_by  VARCHAR(64),
updated_by  VARCHAR(64),
version     INT NOT NULL DEFAULT 0       -- 乐观锁
```

### 命名约束

- 时间字段 `_at` 后缀 + `TIMESTAMPTZ(6)` UTC
- 状态字段 `_status` SMALLINT，禁 VARCHAR ENUM
- 金额字段必带业务前缀：`order_amount` / `fee_amount`，禁裸 `amount`
- 数量字段后缀 `_qty` / `_balance`
- 价格字段后缀 `_price` / `_rate`
- 布尔字段 `is_` 开头：`is_active`、`is_deposit_enabled`

### 索引

- 列出所有索引（主键 + 业务索引 + 外键软索引 + 唯一约束）
- 索引名：`idx_<table>_<col1>_<col2>`，<63 字符
- 大表必须用游标分页索引（`ORDER BY id` + `WHERE id > last_id`）
- **禁数据库级 FOREIGN KEY**（应用层维护关联）

### 注释

每个表 + 关键字段都要 `COMMENT ON COLUMN ...` 说明业务语义。

## 输出格式

```sql
-- [skill: go-team-standards · 数据库设计 · 命名规范] $ARGUMENTS
-- 设计：${author}
-- 创建时间：${now}

CREATE TABLE xxx (
    ...
);

COMMENT ON TABLE xxx IS '...';
COMMENT ON COLUMN xxx.yyy IS '...';

CREATE INDEX idx_xxx_yyy ON xxx(yyy);
```

末尾给一段**字段对照说明**（业务含义、单位、取值范围、是否可空、为什么要这字段），方便 review。

末尾单独一行：🌟
