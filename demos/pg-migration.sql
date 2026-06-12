-- Demo: PostgreSQL Migration — 遵循团队 database.md 规范
-- 文件命名：migrations/000003_create_orders_table.up.sql
-- 配套：migrations/000003_create_orders_table.down.sql（回滚脚本必须成对）
-- 工具：goose 或 golang-migrate 均可

-- ============ up ============

BEGIN;

-- Schema 按业务模块划分，禁止在 public 建业务表
CREATE SCHEMA IF NOT EXISTS trade;

-- 核心业务表：无前缀；表名复数名词
CREATE TABLE trade.orders (
    -- 主键：业务 ID 直接对外暴露（避免 UUID 的索引代价，除非有强需求）
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    -- 逻辑外键（不加数据库级 FOREIGN KEY）
    user_id         BIGINT       NOT NULL,

    -- 金额/数量一律 DECIMAL(28,8)，禁止 FLOAT/DOUBLE
    amount          DECIMAL(28,8) NOT NULL,
    filled_quantity DECIMAL(28,8) NOT NULL DEFAULT 0,

    -- 枚举统一 SMALLINT + 代码常量（0 保留给未知）
    status          SMALLINT     NOT NULL DEFAULT 0,
    side            SMALLINT     NOT NULL,

    -- 幂等键：业务生成，写入保证唯一
    idempotency_key VARCHAR(64)  NOT NULL,

    -- JSONB 仅用于非核心扩展属性
    metadata        JSONB,

    -- 必备字段（审计 + 软删）
    created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ(6),                 -- 软删标记，允许 NULL
    created_by      VARCHAR(64),
    updated_by      VARCHAR(64)
);

-- 注释：生产环境表和关键字段必须 COMMENT
COMMENT ON TABLE  trade.orders               IS '交易订单表，记录用户所有现货/合约订单';
COMMENT ON COLUMN trade.orders.status        IS '订单状态：1=待撮合 2=部分成交 3=完全成交 4=已取消 5=已过期';
COMMENT ON COLUMN trade.orders.side          IS '交易方向：1=买入(BID) 2=卖出(ASK)';
COMMENT ON COLUMN trade.orders.idempotency_key IS '幂等键，由上游业务生成，避免网络重试导致重复下单';

-- 索引：关联字段必建；命名 idx_/uk_/ck_
CREATE INDEX     idx_orders_user_id           ON trade.orders (user_id);
CREATE UNIQUE INDEX uk_orders_idempotency_key ON trade.orders (idempotency_key);
CREATE INDEX     idx_orders_status_created_at ON trade.orders (status, created_at);

-- Check 约束
ALTER TABLE trade.orders
    ADD CONSTRAINT ck_orders_positive_amount CHECK (amount > 0);

COMMIT;


-- ============ down（下面这段放独立的 .down.sql）============
-- BEGIN;
-- DROP TABLE IF EXISTS trade.orders;
-- COMMIT;
