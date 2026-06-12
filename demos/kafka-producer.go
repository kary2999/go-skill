// Demo: Kafka 生产者 — 使用 mask-go-common-lib/mq 与 naming 命名规范
// 禁止直接用 sarama / kafka-go 裸客户端（会绕开命名规范与可观测性）
package demo

import (
	"context"
	"encoding/json"

	"github.com/company/mask-go-common-lib/mq"
	"github.com/company/mask-go-common-lib/naming"
)

// OrderPaidEvent 是业务事件载荷。字段 snake_case，便于跨语言消费。
type OrderPaidEvent struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	AmountRaw string `json:"amount_raw"` // 金额走字符串，避免 JSON 精度丢失
	PaidAt    int64  `json:"paid_at"`    // Unix 毫秒
}

// OrderProducer 用构造注入，禁止全局单例。
type OrderProducer struct {
	p     mq.Producer
	topic string
}

func NewOrderProducer(cfg *mq.Config) (*OrderProducer, func(), error) {
	p, cleanup, err := mq.NewProducer(cfg)
	if err != nil {
		return nil, nil, err
	}
	// 命名规范：{env}_{service}_{entity}_{action}
	topic := naming.TopicInProject(cfg.Env, "order", "pay_status", "updated")
	return &OrderProducer{p: p, topic: topic}, cleanup, nil
}

// PublishPaid 发送订单已支付事件。ctx 必须来自上游（携带 trace/span）。
func (op *OrderProducer) PublishPaid(ctx context.Context, evt *OrderPaidEvent) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return err // 上层 wrap 用 fmt.Errorf("marshal order paid: %w", err)
	}
	// mq.Producer 内部会把 ctx 里的 trace header 注入到 Kafka message headers
	return op.p.Send(ctx, op.topic, evt.OrderID /* partition key */, payload)
}

// 跨项目 topic 示例：
//   topic := naming.TopicCrossProject(cfg.Env, "pay", "order", "record")
//   → prod_pay_to_order_record
