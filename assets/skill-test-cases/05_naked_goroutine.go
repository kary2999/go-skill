//go:build skill_test_only

// skill-test-case: 违反铁律 #5 —— 裸 goroutine
// 期望 AI 指出：go func() 无 ctx / errgroup = 泄漏 + OOM；长期服务 goroutine 数
// 线性增长。改法：errgroup.WithContext 或手动 select ctx.Done()。

package skilltest

import (
	"fmt"
	"time"
)

// ❌ 铁律 #5：没有 ctx，没有退出机制
func StartBackgroundRefresh() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Println("refreshing...") // ❌ 无法被 cancel
		}
	}()
}

// ❌ fan-out 没有 errgroup，错误丢失，goroutine 可能永远 hang
func FanOut(ids []int64) {
	for _, id := range ids {
		go func(id int64) {
			_ = processItem(id) // ❌ 错误被静默吞
		}(id)
	}
}

func processItem(id int64) error {
	return nil
}
