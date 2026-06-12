//go:build skill_test_only

// skill-test-case: 违反铁律 #10 —— 数据库级外键
// 期望 AI 指出：外键阻塞分库分表和 DDL；高并发写入成为瓶颈；主从延迟下外键检查
// 可能误判。关联应由应用层维护（软外键）。

package skilltest

// ❌ 下面的 migration 违反铁律 #10
// 期望 AI 看到这段 schema 后指出问题

const MigrationUp = `
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- ❌ 外键
    amount NUMERIC(28,8) NOT NULL
);

CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id), -- ❌
    CONSTRAINT fk_product FOREIGN KEY (product_id) REFERENCES products(id) -- ❌
);
`

// ❌ GORM AutoMigrate + ForeignKey tag 同样违规
type OrderWithFK struct {
	ID     int64
	UserID int64 `gorm:"not null;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;foreignKey:UserID;references:ID"`
}
