//go:build skill_test_only

// skill-test-case: 违反铁律 #8 —— 用 fmt.Println 打日志
// 期望 AI 指出：非结构化日志无法按字段查；没 trace_id 无法跨服务还原调用链。
// 必须 slog + snake_case 字段 + trace_id / span_id。

package skilltest

import (
	"fmt"
	"log"
)

func HandleOrderCreated(orderID int64, userID int64) {
	// ❌ 铁律 #8：非结构化
	fmt.Printf("order %d created by user %d\n", orderID, userID)
	log.Printf("[INFO] order=%d user=%d status=created", orderID, userID) // ❌ 标准库 log 也不行
}

func OnError(err error) {
	fmt.Println("error:", err) // ❌ 无 level、无 trace_id、无结构
}

// ❌ 拼 JSON 字符串也不行
func Audit(event string, userID int64) {
	fmt.Printf(`{"event":"%s","user":%d}\n`, event, userID)
}
