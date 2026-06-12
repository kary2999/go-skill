//go:build skill_test_only

// skill-test-case: 违反铁律 #9 —— SELECT * / N+1 / OFFSET 分页
// 期望 AI 指出：SELECT * 加字段后破坏协议；循环查询 N+1；OFFSET 大翻页 O(N) 扫描。
// 改法：显式列 + JOIN / IN 批量 + 游标分页（WHERE id > last_id）。

package skilltest

import (
	"database/sql"
	"fmt"
)

type Order struct {
	ID     int64
	UserID int64
}

type User struct {
	ID   int64
	Name string
}

// ❌ SELECT *
func GetOrder(db *sql.DB, id int64) (*Order, error) {
	row := db.QueryRow("SELECT * FROM orders WHERE id = $1", id)
	var o Order
	if err := row.Scan(&o.ID, &o.UserID); err != nil {
		return nil, err
	}
	return &o, nil
}

// ❌ N+1
func ListOrdersWithUser(db *sql.DB, orderIDs []int64) ([]map[string]any, error) {
	var out []map[string]any
	for _, id := range orderIDs {
		var o Order
		_ = db.QueryRow("SELECT id, user_id FROM orders WHERE id = $1", id).Scan(&o.ID, &o.UserID)
		var u User
		_ = db.QueryRow("SELECT id, name FROM users WHERE id = $1", o.UserID).Scan(&u.ID, &u.Name) // ❌ N 次查询
		out = append(out, map[string]any{"order": o, "user": u})
	}
	return out, nil
}

// ❌ OFFSET 翻页，页数大后极慢
func PageOrders(db *sql.DB, page, size int) (*sql.Rows, error) {
	q := fmt.Sprintf("SELECT * FROM orders ORDER BY id LIMIT %d OFFSET %d", size, page*size)
	return db.Query(q)
}
