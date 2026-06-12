// Skill 测试 AI 配置持久化（v1.7.10 起：单一 clawnova provider）
//
// 存储位置（macOS 约定）：
//   ~/Library/Application Support/TeamStandards/eval-config.json
//
// 选这个位置而不是 ~/.team-standards/，是因为：
//   - macOS 下 App 写自己的 Application Support 目录是惯例（已经有 orangecat-template-*.md 在这）
//   - 不进 git（用户 Home 目录与 repo 完全无关）
//   - 退出 App 不丢
//
// 安全：
//   - 文件权限 0600，只本人可读
//   - 仅前端用户主动粘贴时写入；不打日志
//   - 后端 API 返回时把 key 中段星号化（kgb_d0***c244），不直接回显完整 key

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type evalConfig struct {
	Provider string  `json:"provider"`     // 当前固定 "clawnova"，留字段方便未来扩展
	APIBase  string  `json:"api_base"`     // 默认 https://clawnova.ai/api/v1
	APIKey   string  `json:"api_key"`      // 完整 key，本地存储用
	Model    string  `json:"model"`        // 选中的模型，如 qwen3.5-27b-local
	Temperature *float64 `json:"temperature,omitempty"`
}

func evalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	var base string
	switch runtime.GOOS {
	case "darwin":
		base = filepath.Join(home, "Library", "Application Support", "TeamStandards")
	case "windows":
		base = filepath.Join(os.Getenv("APPDATA"), "TeamStandards")
	default:
		base = filepath.Join(home, ".config", "teamstandards")
	}
	return filepath.Join(base, "eval-config.json"), nil
}

func loadEvalConfig() evalConfig {
	// 默认值从 defaults_local.go（gitignored）来。
	// 如果没填 DefaultClawnovaKey（example 模板版），key 是空，用户自己粘。
	cfg := evalConfig{
		Provider: "clawnova",
		APIBase:  DefaultClawnovaAPIBase,
		Model:    DefaultClawnovaModel,
		APIKey:   DefaultClawnovaKey, // 默认 key（可能为空）
	}
	p, err := evalConfigPath()
	if err != nil {
		return cfg
	}
	b, err := os.ReadFile(p)
	if err != nil {
		// 文件不存在 → 返回默认（含默认 key），相当于"开箱即用"
		return cfg
	}
	var saved evalConfig
	if err := json.Unmarshal(b, &saved); err != nil {
		return cfg
	}
	if saved.Provider != "" {
		cfg.Provider = saved.Provider
	}
	if saved.APIBase != "" {
		cfg.APIBase = saved.APIBase
	}
	if saved.APIKey != "" {
		cfg.APIKey = saved.APIKey
	}
	if saved.Model != "" {
		cfg.Model = saved.Model
	}
	if saved.Temperature != nil {
		cfg.Temperature = saved.Temperature
	}
	return cfg
}

func saveEvalConfig(cfg evalConfig) error {
	p, err := evalConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(p, b, 0600)
}

// maskKey 中段星号化 kgb_d0ef5b3...c12190e0127120ae244f → kgb_d0ef***ae244f
func maskKey(k string) string {
	if k == "" {
		return ""
	}
	if len(k) <= 12 {
		return strings.Repeat("*", len(k))
	}
	return k[:8] + "***" + k[len(k)-6:]
}

// GET /api/eval/config —— 返回当前配置（key 星号化）+ 健康标记
func handleEvalConfigGet(w http.ResponseWriter, r *http.Request) {
	cfg := loadEvalConfig()

	// 检测明显坏的 key（v1.7.12 起，避免 1.7.11 之前用户存了 URL 当 key 的脏数据）
	keyHealth := "ok"
	if cfg.APIKey == "" {
		keyHealth = "missing"
	} else if strings.HasPrefix(cfg.APIKey, "http://") || strings.HasPrefix(cfg.APIKey, "https://") {
		keyHealth = "bad_is_url" // 上次存的是 URL（误填），用户应清除已存重新来
	} else if len(cfg.APIKey) < 16 {
		keyHealth = "bad_too_short"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"provider":    cfg.Provider,
		"api_base":    cfg.APIBase,
		"model":       cfg.Model,
		"temperature": cfg.Temperature,
		"key_masked":  maskKey(cfg.APIKey),
		"has_key":     cfg.APIKey != "",
		"key_health":  keyHealth,
		"path":        func() string { p, _ := evalConfigPath(); return p }(),
	})
}

// POST /api/eval/config —— 保存配置
func handleEvalConfigSave(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body", err)
		return
	}
	var req struct {
		APIBase string  `json:"api_base"`
		APIKey  string  `json:"api_key"` // 空 = 不变；非空 = 覆盖
		Model   string  `json:"model"`
		Temperature *float64 `json:"temperature,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err)
		return
	}

	cfg := loadEvalConfig()
	if req.APIBase != "" {
		cfg.APIBase = req.APIBase
	}
	if req.APIKey != "" {
		// 防呆：key 不能是 URL（最常见的填错框场景）
		k := strings.TrimSpace(req.APIKey)
		if strings.HasPrefix(k, "http://") || strings.HasPrefix(k, "https://") {
			writeError(w, http.StatusBadRequest,
				"看起来你把 URL 粘到了 API Key 框（key 不该以 http:// 或 https:// 开头）。"+
					"先把 URL 粘到「API Base」框，把真正的 kgb_... key 粘到「API Key」框，再点保存。",
				nil)
			return
		}
		// 太短肯定不是真 key
		if len(k) < 16 {
			writeError(w, http.StatusBadRequest,
				"API Key 太短（少于 16 字符），看起来不是完整的 key。请粘完整的 kgb_... 字符串。", nil)
			return
		}
		cfg.APIKey = k
	}
	if req.Model != "" {
		cfg.Model = req.Model
	}
	if req.Temperature != nil {
		cfg.Temperature = req.Temperature
	}

	if err := saveEvalConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "save config", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"key_masked": maskKey(cfg.APIKey),
		"has_key":    cfg.APIKey != "",
	})
}

// DELETE /api/eval/config —— 清除 key（保留 base/model 选择）
func handleEvalConfigClearKey(w http.ResponseWriter, r *http.Request) {
	cfg := loadEvalConfig()
	cfg.APIKey = ""
	if err := saveEvalConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "save", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /api/eval/test-connection —— 用当前配置发一个最小请求验证连通性
func handleEvalTestConnection(w http.ResponseWriter, r *http.Request) {
	cfg := loadEvalConfig()
	if cfg.APIKey == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":    false,
			"error": "未配置 API Key",
		})
		return
	}

	pcfg := ProviderConfig{
		Provider: cfg.Provider,
		APIBase:  cfg.APIBase,
		APIKey:   cfg.APIKey,
		Model:    cfg.Model,
	}
	res := callByProvider(pcfg, "你是测试助手。", "回复字符串：OK")

	if res.RawError != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":    false,
			"error": res.RawError.Error(),
			"model": cfg.Model,
			"base":  cfg.APIBase,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"model":         cfg.Model,
		"base":          cfg.APIBase,
		"reply":         res.Text,
		"input_tokens":  res.InputTokens,
		"output_tokens": res.OutputTokens,
	})
}
