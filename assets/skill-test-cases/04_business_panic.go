//go:build skill_test_only

// skill-test-case: 违反铁律 #4 —— 业务代码 panic
// 期望 AI 指出：panic 挂整进程，K8s 会滚动重启，影响实例上所有请求；只有 init()
// 或启动阶段才允许 panic。改法：返回 error 向上冒泡。

package skilltest

import "fmt"

type Order struct {
	ID     int64
	Amount int64
}

func (o *Order) Validate() {
	if o.ID <= 0 {
		panic("invalid order id") // ❌ 业务代码 panic
	}
	if o.Amount < 0 {
		panic(fmt.Sprintf("negative amount: %d", o.Amount)) // ❌
	}
}

func MustLoadFromDB(id int64) *Order {
	if id == 0 {
		panic("id required") // ❌ Must* 命名也不该在请求路径用
	}
	return &Order{ID: id}
}
