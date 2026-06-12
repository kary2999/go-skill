// Demo: Kafka 消费者 — 使用 mask-go-common-lib/mq，带优雅退出
package demo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/company/mask-go-common-lib/logging"
	"github.com/company/mask-go-common-lib/mq"
)

type OrderPaidHandler struct {
	log *logging.Logger
	// 注入下游 usecase，消费者层不做业务
	uc OrderUsecase
}

type OrderUsecase interface {
	MarkOrderPaid(ctx context.Context, orderID string) error
}

// Handle 是消费回调；返回 error 表示处理失败，mq 层决定是否重试 / 进死信队列。
func (h *OrderPaidHandler) Handle(ctx context.Context, msg *mq.Message) error {
	var evt OrderPaidEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		// 协议错误一般不可恢复，落告警队列即可，不要无限重试
		h.log.ErrorContext(ctx, "invalid order paid payload",
			"topic", msg.Topic, "err", err)
		return nil // 不重试
	}

	if err := h.uc.MarkOrderPaid(ctx, evt.OrderID); err != nil {
		return fmt.Errorf("mark order paid %s: %w", evt.OrderID, err)
	}
	return nil
}

// StartConsumer 在 cmd/main.go 或 server 初始化时调用。
func StartConsumer(ctx context.Context, cfg *mq.Config, h *OrderPaidHandler) error {
	c, err := mq.NewConsumer(cfg, &mq.ConsumerOptions{
		Topic:   cfg.Topics["order_pay_status_updated"], // 来自 config 的命名映射
		GroupID: cfg.Env + "_order_pay_status_consumer",
		Handler: h.Handle,
	})
	if err != nil {
		return err
	}
	// Start 会阻塞直到 ctx 取消（优雅退出）
	return c.Start(ctx)
}
