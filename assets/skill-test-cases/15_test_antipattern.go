//go:build skill_test_only

// skill-test-case: 单元测试反模式（对应 go-unit-test Skill）
// 期望 AI 指出多个问题：
//   1. 手写 mock struct（应 mockgen）
//   2. 直接调 time.Now()（应注入 Clock）
//   3. err.Error() 字符串比较（应 errors.Is）
//   4. 共享全局 mock（应 case 内独立）
//   5. 非表驱动 + 非 t.Run
//   6. 单测里直连真 DB（应 sqlmock 或集成测试）

package skilltest_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"
)

// ❌ 手写 mock，接口改了忘改 mock 就假绿
type fakeOrderRepo struct {
	Calls int
}

func (f *fakeOrderRepo) Create(o any) (int64, error) {
	f.Calls++
	return 1, nil
}

// ❌ 全局共享
var sharedRepo = &fakeOrderRepo{}

func TestCreateOrder_Old(t *testing.T) {
	// ❌ 不是表驱动，case 全堆一起

	// ❌ 直接调 time.Now()，测试不可重放
	now := time.Now()
	_ = now

	// ❌ 直连真 DB
	db, _ := sql.Open("postgres", "postgres://localhost/prod")
	defer db.Close()

	// 场景 1
	_, err := sharedRepo.Create(map[string]any{"amount": 100})
	if err != nil {
		// ❌ 字符串比较错误
		if strings.Contains(err.Error(), "not found") {
			t.Fatal("got not found")
		}
		// 应该：errors.Is(err, ErrNotFound)
	}

	// 场景 2（应该拆到另一个 case）
	_, err = sharedRepo.Create(nil)
	if err != nil && err.Error() == "invalid argument" { // ❌ 文案比较
		t.Log("ok")
	}

	// ❌ 没 defer ctrl.Finish() / 没清理
	_ = errors.New("dummy")
}
