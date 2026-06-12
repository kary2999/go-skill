// Demo: slog + OTel — 结构化日志自动携带 trace_id / span_id
// 团队包 mask-go-common-lib/logging 已做封装；下面演示直接使用 slog 时如何手工提取
package service

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/company/mask-go-common-lib/logging"
)

// HandleCreateOrder 是 service 层方法，ctx 由 middleware 注入 trace context
func (s *OrderService) HandleCreateOrder(ctx context.Context, req *CreateOrderReq) error {
	// ✓ 推荐：直接用 logging.New 返回的 *slog.Logger，自动从 ctx 提取 trace 字段
	s.log.InfoContext(ctx, "create order start",
		slog.String("user_id", req.UserID),
		slog.String("idempotency_key", req.IdempotencyKey),
		slog.String("amount_raw", req.AmountRaw), // 金额别转 float64，直接传原串
	)

	if err := s.uc.CreateOrder(ctx, req); err != nil {
		// error 级别日志：携带 error_code 让 Loki 能按码聚合
		s.log.ErrorContext(ctx, "create order failed",
			slog.String("user_id", req.UserID),
			slog.String("err", err.Error()),
			slog.Int64("error_code", extractErrorCode(err)),
		)
		return err
	}

	s.log.InfoContext(ctx, "create order success",
		slog.String("user_id", req.UserID),
	)
	return nil
}

// 手工提取 trace_id / span_id 的示例（通常不用，middleware 已做）
func logFieldsFromCtx(ctx context.Context) []slog.Attr {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil
	}
	return []slog.Attr{
		slog.String("trace_id", sc.TraceID().String()),
		slog.String("span_id", sc.SpanID().String()),
	}
}

// ----- 严禁写法 -----
// func badLogging() {
//     fmt.Println("created order: " + orderID)          // ✗ 非结构化
//     log.Printf("user=%s amount=%f", userID, amount)   // ✗ 字符串拼接 + float
// }

type OrderService struct {
	log *logging.Logger
	uc  OrderUsecase
}

type OrderUsecase interface {
	CreateOrder(ctx context.Context, req *CreateOrderReq) error
}

type CreateOrderReq struct {
	UserID         string
	IdempotencyKey string
	AmountRaw      string
}

func extractErrorCode(err error) int64 { return 0 /* xerror.As 提取 */ }
