// Demo: GORM Repo 模式 — 游标分页、软删、context 超时、禁止 SELECT *
package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/company/mask-go-common-lib/logging"
)

// Order 对应 trade.orders 表；字段顺序与表一致便于审查
type Order struct {
	ID             int64           `gorm:"primaryKey"`
	UserID         int64           `gorm:"column:user_id"`
	Amount         decimal.Decimal `gorm:"type:decimal(28,8)"`
	FilledQuantity decimal.Decimal `gorm:"type:decimal(28,8)"`
	Status         int16
	Side           int16
	IdempotencyKey string `gorm:"column:idempotency_key"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"` // 软删
}

func (Order) TableName() string { return "trade.orders" }

// OrderRepo 定义契约（接口放使用方，实现放 data 包）
// 使用方（biz 层）依赖这个接口而非 *OrderRepoImpl
type OrderRepo interface {
	Create(ctx context.Context, o *Order) (*Order, error)
	GetByID(ctx context.Context, id int64) (*Order, error)
	ListByUserCursor(ctx context.Context, userID, afterID int64, limit int) ([]*Order, error)
}

type orderRepo struct {
	db  *gorm.DB
	log *logging.Logger
}

func NewOrderRepo(db *gorm.DB, log *logging.Logger) OrderRepo {
	return &orderRepo{db: db, log: log}
}

// Create 写入订单；幂等键冲突时返回已有记录而非错误（契约）。
func (r *orderRepo) Create(ctx context.Context, o *Order) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second) // DB 查询必须带超时
	defer cancel()

	err := r.db.WithContext(ctx).Create(o).Error
	if err != nil {
		// 实际生产会区分唯一冲突 → 走幂等分支
		return nil, fmt.Errorf("create order (idem=%s): %w", o.IdempotencyKey, err)
	}
	return o, nil
}

// GetByID 软删记录自动排除。
func (r *orderRepo) GetByID(ctx context.Context, id int64) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	var o Order
	err := r.db.WithContext(ctx).
		Select("id, user_id, amount, filled_quantity, status, side, idempotency_key, created_at, updated_at").
		Where("id = ?", id).
		First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order %d: %w", id, err)
	}
	return &o, nil
}

// ListByUserCursor 使用游标分页，禁止大 OFFSET。
// afterID = 0 表示第一页；调用方拿到最后一条的 ID 传回作下一页的 afterID。
func (r *orderRepo) ListByUserCursor(ctx context.Context, userID, afterID int64, limit int) ([]*Order, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	var out []*Order
	err := r.db.WithContext(ctx).
		Select("id, user_id, amount, status, created_at").
		Where("user_id = ? AND id > ?", userID, afterID).
		Order("id ASC").
		Limit(limit).
		Find(&out).Error
	if err != nil {
		return nil, fmt.Errorf("list orders user=%d after=%d: %w", userID, afterID, err)
	}
	return out, nil
}

// ErrOrderNotFound 应放在 biz 或 errno 包（业务错误走 xerror.New(errno.OrderNotFound)）
var ErrOrderNotFound = errors.New("order not found")
