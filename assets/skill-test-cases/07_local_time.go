//go:build skill_test_only

// skill-test-case: 违反铁律 #7 —— 时间不带时区 / 本地时间
// 期望 AI 指出：跨区 + 跨 DC 部署下本地时间边界会产生一天差；夏令时切换还会重复/消失。
// 必须 UTC 存 TIMESTAMPTZ(6)，对外 ISO 8601 含时区。

package skilltest

import (
	"fmt"
	"time"
)

type Event struct {
	ID        int64
	CreatedAt time.Time // 用法不对：下面存入时用了 time.Now() 本地
}

// ❌ time.Now() 是本地时区，存库后别的 DC 读出来会解释成它的本地时区
func NewEvent(id int64) *Event {
	return &Event{
		ID:        id,
		CreatedAt: time.Now(),
	}
}

// ❌ 日期字符串不带时区信息
func FormatCreatedAt(e *Event) string {
	return e.CreatedAt.Format("2006-01-02 15:04:05") // ❌ 无 tz
}

// ❌ 解析用户输入时假定是本地时间
func ParseUserDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func Demo() {
	t, _ := ParseUserDate("2026-04-23")
	fmt.Println(t.Location()) // UTC，但用户可能期望他的本地
}
