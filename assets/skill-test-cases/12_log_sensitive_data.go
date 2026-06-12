//go:build skill_test_only

// skill-test-case: 违反铁律 #12 —— 敏感数据进日志 / prompt
// 期望 AI 指出：手机号、身份证、邮箱、token 进 Loki / ES 长期存储 → GDPR 合规事故。
// 必须脱敏（*** 中间位）或不打印。

package skilltest

import (
	"fmt"
	"log/slog"
)

type User struct {
	ID       int64
	Phone    string // 13800000000
	Email    string
	IDCard   string // 110101199001011234
	BankCard string // 6222020200000000000
	APIToken string
}

func DebugLogin(u *User) {
	// ❌ 全打进日志
	slog.Info("user login",
		"user_id", u.ID,
		"phone", u.Phone, // ❌
		"email", u.Email, // ❌
		"id_card", u.IDCard, // ❌
		"token", u.APIToken, // ❌
	)
}

// ❌ 打印完整银行卡
func AuditTransfer(u *User, amount int64) {
	fmt.Printf("user=%d bank_card=%s amount=%d\n", u.ID, u.BankCard, amount)
}

// ❌ 写入测试 fixture / AI prompt
func BuildAIPrompt(u *User) string {
	return fmt.Sprintf("help me analyze user %s (phone: %s, id: %s)",
		u.Email, u.Phone, u.IDCard)
}
