// Demo: errno + xerror — 统一错误码使用
// 注意：errno 包的常量由 IDP 错误码仓库 codegen 生成，禁止手写数字字面量
package biz

import (
	"context"
	"errors"
	"fmt"

	"github.com/company/mask-go-common-lib/errors/xerror"
	"github.com/company/pkg/errno" // ← IDP codegen 产物（每个服务引用）
)

type OrderUsecase struct {
	repo   OrderRepository
	locker DistLocker
}

type OrderRepository interface {
	GetByIdempotencyKey(ctx context.Context, key string) (*Order, error)
	Create(ctx context.Context, o *Order) (*Order, error)
}

type DistLocker interface {
	Acquire(ctx context.Context, key string) (release func(), err error)
}

type Order struct {
	ID             int64
	IdempotencyKey string
	// ...
}

// CreateOrder 返回的 error 必须是 *xerror.Error（或被 fmt.Errorf 包装），
// 上层 handler 会调用 xerror.As 提取 code 塞到响应 + 日志 tag。
func (uc *OrderUsecase) CreateOrder(ctx context.Context, req *CreateOrderReq) (*Order, error) {
	if req.IdempotencyKey == "" {
		// ✗ 禁止：return errors.New("idempotency key required")
		// ✗ 禁止：return xerror.New(400, "idempotency key required")
		// ✓ 标准写法：
		return nil, xerror.New(errno.InvalidArgument).
			WithDetail("idempotency_key is required")
	}

	// 幂等检查
	existing, err := uc.repo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil && !errors.Is(err, ErrNotFound) {
		// 基础设施错误用 %w 包装，保留链式信息
		return nil, fmt.Errorf("get by idem key: %w", err)
	}
	if existing != nil {
		return existing, nil // 幂等命中，契约里已约定返回已有记录而非 error
	}

	// 拿分布式锁避免并发重复创建
	release, err := uc.locker.Acquire(ctx, "order:create:"+req.IdempotencyKey)
	if err != nil {
		return nil, xerror.New(errno.ResourceBusy).
			WithDetail("order creation is locked, retry later")
	}
	defer release()

	order, err := uc.repo.Create(ctx, newOrderFromReq(req))
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return order, nil
}

type CreateOrderReq struct {
	IdempotencyKey string
	// ...
}

var ErrNotFound = errors.New("not found")

func newOrderFromReq(req *CreateOrderReq) *Order {
	return &Order{IdempotencyKey: req.IdempotencyKey}
}

// === IDP codegen 产物的引用方式（仅示意） ===
// package errno
//
// const (
//     InvalidArgument  = 100400
//     ResourceBusy     = 100429
//     OrderNotFound    = 100404
// )
//
// 实际产物会带 String() / HTTPStatus() 方法，由 codegen 流水线自动生成。
