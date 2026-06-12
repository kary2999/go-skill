//go:build skill_test_only

// skill-test-case: 违反铁律 #1 —— 硬编码密钥
// 期望 AI 指出：API Key / DB 密码不应写在源码里，应走 env / config，并说明
// "git 历史永久记录，一旦入库就泄密"。给出修正代码（config.Load + env 注入）。

package skilltest

import (
	"fmt"
	"net/http"
)

const (
	// ❌ 铁律 #1：禁硬编码密钥
	APIKey       = "sk-live-9f3e2a7b6c1d0f8e5a4b3c2d1e0f9a8b"
	DBPassword   = "prod-pg-password-20260101!"
	InternalHMAC = "hmac-secret-do-not-share-in-log"
)

func doCall() {
	req, _ := http.NewRequest("GET", "https://api.example.com/v1/balance", nil)
	req.Header.Set("Authorization", "Bearer "+APIKey) // ❌ 密钥在 binary 里
	fmt.Println(DBPassword)
}
