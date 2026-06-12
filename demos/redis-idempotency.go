// Demo: Redis 幂等键 + 分布式锁 — 使用 mask-go-common-lib/redisx
// 禁止绕过 redisx 直连 go-redis（会绕开命名规范）
package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/company/mask-go-common-lib/redisx"
)

// IdempotencyStore 语义化封装，业务层只看到"检查幂等 / 记录幂等"两个动作
type IdempotencyStore struct {
	store redisx.Store
}

func NewIdempotencyStore(s redisx.Store) *IdempotencyStore {
	return &IdempotencyStore{store: s}
}

// CheckAndReserve 原子检查 + 占位：
//   - 首次调用返回 (false, nil)，表示"你是第一个，继续往下执行业务"
//   - 重复调用返回 (true, nil)，调用方应读取缓存结果直接返回
// Key 命名走 redisx 统一规范：{业务}:{模块}:{唯一标识}
func (s *IdempotencyStore) CheckAndReserve(ctx context.Context, bizDomain, module, key string, ttl time.Duration) (bool, error) {
	ok, err := s.store.SetNX(ctx, bizDomain, module, key, "1", ttl)
	if err != nil {
		return false, fmt.Errorf("idem setnx %s:%s:%s: %w", bizDomain, module, key, err)
	}
	// ok=true 表示我们是写入者，即首次；ok=false 表示已存在，即重复请求
	return !ok, nil
}

// --- 分布式锁 ---

// DistLock 基于 SET NX PX + lua 解锁。锁名也走命名规范。
type DistLock struct {
	store redisx.Store
}

func NewDistLock(s redisx.Store) *DistLock { return &DistLock{store: s} }

// Acquire 返回 token（释放凭证）；token 为空表示未拿到锁。
func (l *DistLock) Acquire(ctx context.Context, bizDomain, module, key string, ttl time.Duration) (string, error) {
	token := genToken() // uuid / snowflake
	ok, err := l.store.SetNX(ctx, bizDomain, module, "lock:"+key, token, ttl)
	if err != nil {
		return "", fmt.Errorf("dist lock acquire: %w", err)
	}
	if !ok {
		return "", nil
	}
	return token, nil
}

// Release 使用 lua 原子解锁，避免误删他人持有的锁。
const releaseLuaScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end`

func (l *DistLock) Release(ctx context.Context, bizDomain, module, key, token string) error {
	n, err := l.store.EvalInt(ctx, releaseLuaScript, []string{
		redisx.BuildKey(bizDomain, module, "lock:"+key),
	}, token)
	if err != nil {
		return fmt.Errorf("dist lock release: %w", err)
	}
	if n == 0 {
		return errors.New("lock token mismatch (already expired or stolen)")
	}
	return nil
}

// genToken 应使用 crypto/rand 生成，略。
func genToken() string { return "TODO-replace-with-crypto/rand-uuid" }
