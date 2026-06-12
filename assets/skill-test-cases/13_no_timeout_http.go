//go:build skill_test_only

// skill-test-case: 违反铁律 #13 —— 外部 IO 无超时 / 无 ctx / trace 断链
// 期望 AI 指出：上游慢调用穿透 → 雪崩；无 ctx = 分布式追踪失效；标准库 http.Client 零值
// 无超时。必须用 httpclient.New（common-lib）或显式设置 Timeout + 传 ctx。

package skilltest

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
)

// ❌ 默认 http.Client 没有超时
func FetchUser(userID int64) (map[string]any, error) {
	resp, err := http.Get("https://api.example.com/users/" + itoa(userID)) // ❌ 无 ctx、无超时
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var m map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&m)
	return m, nil
}

// ❌ 自建 client 但没设 Timeout
func upstreamCall() *http.Client {
	return &http.Client{} // ❌
}

// ❌ DB 查询也没传 ctx
func GetOrder(db *sql.DB, id int64) error {
	_, err := db.Exec("UPDATE orders SET status='paid' WHERE id=$1", id) // ❌ 应 ExecContext(ctx, ...)
	return err
}

// ❌ 接收到 ctx 却没往下传
func HandleRequest(ctx context.Context) {
	http.Get("https://slow.internal/data") // ❌ ctx 就地丢弃
}

func itoa(n int64) string { return "" }
