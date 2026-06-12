package main

// team-guard 安装时自动写入 Claude Code settings.json
// 需要将同一脚本注册到三个 hook 事件：UserPromptSubmit / PreToolUse / PostToolUse

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// resolveSettingsFile 根据 scope 返回 settings.json 路径
//   global  → ~/.claude/settings.json
//   project → <projectPath>/.claude/settings.json
func resolveSettingsFile(scope, projectPath string) (string, error) {
	switch scope {
	case "project":
		if projectPath == "" {
			return "", fmt.Errorf("scope=project 时 path 必填")
		}
		abs, err := filepath.Abs(projectPath)
		if err != nil {
			return "", err
		}
		return filepath.Join(abs, ".claude", "settings.json"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude", "settings.json"), nil
	}
}

// teamGuardHookEntry 是 settings.json hooks 数组里的一项
type teamGuardHookEntry struct {
	Matcher string              `json:"matcher,omitempty"`
	Hooks   []teamGuardHookCmd  `json:"hooks"`
}
type teamGuardHookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// injectTeamGuardSettings 将 team-guard.sh 注册到 settings.json 的三个事件
// 已存在则幂等（不重复添加）
func injectTeamGuardSettings(scope, projectPath, scriptPath string) error {
	settingsPath, err := resolveSettingsFile(scope, projectPath)
	if err != nil {
		return err
	}

	// 读现有 settings（可能不存在）
	raw := map[string]any{}
	if b, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(b, &raw)
	}

	cmd := "bash " + scriptPath

	// 确保 hooks 节点存在
	hooksNode, _ := raw["hooks"].(map[string]any)
	if hooksNode == nil {
		hooksNode = map[string]any{}
	}

	// 三个事件的配置
	type eventCfg struct {
		event   string
		matcher string
	}
	events := []eventCfg{
		{"UserPromptSubmit", ""},
		{"PreToolUse", "Bash"},
		{"PostToolUse", "Write|Edit"},
	}

	for _, ec := range events {
		arr, _ := hooksNode[ec.event].([]any)
		// 检查是否已经注册了 team-guard
		alreadyIn := false
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			hooks, _ := m["hooks"].([]any)
			for _, h := range hooks {
				hm, ok := h.(map[string]any)
				if !ok {
					continue
				}
				if hm["command"] == cmd {
					alreadyIn = true
					break
				}
			}
			if alreadyIn {
				break
			}
		}
		if alreadyIn {
			continue
		}

		entry := map[string]any{
			"hooks": []any{
				map[string]any{"type": "command", "command": cmd},
			},
		}
		if ec.matcher != "" {
			entry["matcher"] = ec.matcher
		}
		arr = append(arr, entry)
		hooksNode[ec.event] = arr
	}

	raw["hooks"] = hooksNode

	// 写回（缩进 2 空格，保持可读）
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, b, 0644)
}

// removeTeamGuardSettings 从 settings.json 移除 team-guard 的三个事件注册
func removeTeamGuardSettings(scope, projectPath, scriptPath string) error {
	settingsPath, err := resolveSettingsFile(scope, projectPath)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil // 文件不存在，无需处理
	}
	raw := map[string]any{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil
	}

	cmd := "bash " + scriptPath
	hooksNode, _ := raw["hooks"].(map[string]any)
	if hooksNode == nil {
		return nil
	}

	for event, val := range hooksNode {
		arr, _ := val.([]any)
		filtered := make([]any, 0, len(arr))
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				filtered = append(filtered, item)
				continue
			}
			hooks, _ := m["hooks"].([]any)
			newHooks := make([]any, 0, len(hooks))
			for _, h := range hooks {
				hm, ok := h.(map[string]any)
				if ok && hm["command"] == cmd {
					continue // 跳过 team-guard 项
				}
				newHooks = append(newHooks, h)
			}
			if len(newHooks) > 0 {
				m["hooks"] = newHooks
				filtered = append(filtered, m)
			}
			// 若 hooks 全部清空则整条 entry 丢弃
		}
		if len(filtered) > 0 {
			hooksNode[event] = filtered
		} else {
			delete(hooksNode, event)
		}
	}

	raw["hooks"] = hooksNode
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0644)
}
