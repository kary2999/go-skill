//go:build skill_test_only

// skill-test-case: 违反铁律 #14 —— Prometheus metric label 高基数
// 期望 AI 指出：user_id / trace_id / order_id 作 label → series 指数爆炸 →
// Prometheus 内存磁盘爆掉。改法：只用低基数 label（status、method、route 模板），
// 把业务 ID 放 trace / log，不要放 metric。

package skilltest

import "github.com/prometheus/client_golang/prometheus"

// ❌ user_id 作 label：每个用户都是一条 series
var UserRequestCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "user_requests_total"},
	[]string{"user_id", "method"}, // ❌
)

// ❌ order_id + trace_id 双高基数
var OrderLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{Name: "order_latency_seconds"},
	[]string{"order_id", "trace_id", "user_id"}, // ❌❌❌
)

// ❌ URL path 不做路由模板化，每个 uuid 一条 series
var HTTPRequestCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "http_requests_total"},
	[]string{"path"}, // ❌ 应用模板 /users/:id，不是原始 /users/abc-123
)

// ❌ 错误消息文本作 label（无穷可能）
var ErrorCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "errors_total"},
	[]string{"error_message"}, // ❌
)
