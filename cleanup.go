// 启动时自动清理旧版本（≤ 1.3.0）残留的磁盘痕迹。
// 这些旧 tab / 模块已被删除（prompt-rules, market, find-skills-listing），
// 但用户之前装过的话会在系统里留下：
//   - ~/.claude/CLAUDE.md 里的 <!-- TEAM_STANDARDS_PROMPT_BEGIN --> 段
//   - ~/.cursor/rules/00-attitude.mdc
//   - ~/.team-standards/prompt-rules.json
// 这些会和现在唯一保留的 persona 段并存，看起来像"重复"。
package main

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	legacyPromptBegin = "<!-- TEAM_STANDARDS_PROMPT_BEGIN -->"
	legacyPromptEnd   = "<!-- TEAM_STANDARDS_PROMPT_END -->"
)

func cleanupLegacyArtifacts() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// 1. 从 ~/.claude/CLAUDE.md 移除旧 PROMPT 段（persona 段不动）
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	if b, err := os.ReadFile(claudeMD); err == nil {
		s := string(b)
		i := strings.Index(s, legacyPromptBegin)
		j := strings.Index(s, legacyPromptEnd)
		if i >= 0 && j > i {
			out := strings.TrimRight(s[:i], "\n ") + "\n" +
				strings.TrimLeft(s[j+len(legacyPromptEnd):], "\n ")
			out = strings.TrimSpace(out)
			if out == "" {
				_ = os.Remove(claudeMD)
			} else {
				_ = os.WriteFile(claudeMD, []byte(out+"\n"), 0644)
			}
			logInfo("cleanup", "已移除 CLAUDE.md 里旧 TEAM_STANDARDS_PROMPT 段（v1.3.0 遗留）")
		}
	}

	// 2. 删旧的 00-attitude.mdc（persona 现在走 00-persona.mdc，attitude 是旧文件名）
	if p := filepath.Join(home, ".cursor", "rules", "00-attitude.mdc"); remove(p) {
		logInfo("cleanup", "已删 "+p)
	}

	// 3. 删旧的 prompt-rules.json 存储
	if p := filepath.Join(home, ".team-standards", "prompt-rules.json"); remove(p) {
		logInfo("cleanup", "已删 "+p)
	}
}

func remove(p string) bool {
	if _, err := os.Stat(p); err == nil {
		return os.Remove(p) == nil
	}
	return false
}
