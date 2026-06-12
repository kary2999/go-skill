//go:build skill_test_only

// skill-test-case: 违反铁律 #6 —— 金额用 float64
// 期望 AI 指出：浮点精度丢失，0.1+0.2 != 0.3，账务对不平；必须用 decimal.Decimal
// + DB DECIMAL(28,8)。

package skilltest

import "fmt"

type Wallet struct {
	UserID  int64
	Balance float64 // ❌ 铁律 #6：金额用 float64
}

// ❌ 浮点累加必然精度丢失
func (w *Wallet) Deposit(amount float64) {
	w.Balance += amount
}

// ❌ 折扣计算 + 浮点比较必炸
func FinalPrice(original, discount float64) float64 {
	final := original * (1.0 - discount)
	if final == 0.0 { // ❌ float 禁止 == 比较
		return 0
	}
	return final
}

func Demo() {
	a := 0.1
	b := 0.2
	fmt.Println(a + b) // 0.30000000000000004
}
