//go:build skill_test_only

// skill-test-case: 违反铁律 #3 —— error 必检 + 包装
// 期望 AI 指出：_ = err / 忽略 err 让事故无线索；跨层必须 fmt.Errorf("ctx: %w", err)
// 保留错误链，上层才能 errors.Is / errors.As。

package skilltest

import (
	"encoding/json"
	"os"
)

type Config struct {
	Env string `json:"env"`
}

func LoadConfig() *Config {
	b, _ := os.ReadFile("/etc/app/config.json") // ❌ 丢了 err
	var c Config
	_ = json.Unmarshal(b, &c) // ❌ 又丢了一次
	return &c
}

func SaveState(state string) {
	f, err := os.Create("/tmp/state.txt")
	if err != nil {
		return // ❌ 上层无从得知失败
	}
	_, _ = f.WriteString(state) // ❌ 写入错误没人处理
	_ = f.Close()
}
