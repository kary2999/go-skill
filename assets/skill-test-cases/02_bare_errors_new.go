//go:build skill_test_only

// skill-test-case: 违反铁律 #2 —— 错误未走 xerror + errno
// 期望 AI 指出：业务错误必须用 xerror.New(errno.XxxCode)，否则 Loki 无法按 code
// 聚合、前端无法国际化。给出 errno 定义 + xerror.New 修正。

package skilltest

import (
	"errors"
	"fmt"
)

// ❌ 铁律 #2：裸 errors.New 没有 code，监控聚合失败
var ErrOrderNotFound = errors.New("order not found")
var ErrPaymentFailed = errors.New("payment failed: downstream rejected")

func GetOrder(id int64) error {
	if id <= 0 {
		return errors.New("invalid id") // ❌ 没错误码
	}
	// 模拟查库
	return fmt.Errorf("db error: %w", ErrOrderNotFound) // 包装了但里面还是裸字符串
}
