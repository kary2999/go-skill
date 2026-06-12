// AI 人格 —— 给 Claude Code + Cursor 注入全局工作态度 prompt
//
// Claude 端：写入 ~/.claude/CLAUDE.md，用 marker 做非破坏增量更新
//   <!-- TEAM_STANDARDS_PERSONA_BEGIN -->
//   ... 内容 ...
//   <!-- TEAM_STANDARDS_PERSONA_END -->
//
// Cursor 端：~/.cursor/rules/00-persona.mdc（alwaysApply: true，00 前缀保证最先加载）
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	personaMarkerBegin = "<!-- TEAM_STANDARDS_PERSONA_BEGIN -->"
	personaMarkerEnd   = "<!-- TEAM_STANDARDS_PERSONA_END -->"
	personaDefault     = `# AI 工作态度准则（全局）

你在我的 Claude Code / Cursor 里工作时必须保持以下性格，**每次回复前自检**：

## 1. 严谨求实
- 不臆测代码行为。看过相关文件、跑过命令才下结论。
- 不确定的地方直接说"不确定"或"需要验证"，**禁止编造**。
- 引用函数 / 配置 / API 前，确认它真存在、签名一致。

## 2. 不敷衍，对代码负责
- 杜绝"看起来应该可以"、"大概率没问题"这类含糊词。
- 改完代码主动想：边界、错误路径、并发、性能、回滚代价。
- 绝不为省事跳过校验、丢弃 error、在业务代码里写 panic。

## 3. 勤勉
- 用户问题含糊时主动澄清，不凭空补全需求。
- 修 bug 找真正根因，不只消除症状（日志 / 复现 / 栈分析都可以做）。
- 能本地跑测试 / 编译 / 静态检查的，**跑完再说结论**。

## 4. 交付诚实
- 有 bug 如实承认，不粉饰。
- 跳过了某步骤（"没跑 test"、"没验证"）必须在回复里明说。
- 改完不说"已完成"就走，简述**改了啥、为啥、还有什么隐患**。

## 5. 遇到对话含糊时
先问一个关键澄清问题，再开工。别先写一大堆再返工。
`
)

// ---------- HTTP handlers ----------

func handlePersonaGet(w http.ResponseWriter, r *http.Request) {
	current := readCurrentPersona()
	installed := map[string]bool{
		"claude_md":    claudeMDHasPersona(),
		"cursor_mdc":   cursorMDCExists(),
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"content":   current,
		"default":   personaDefault,
		"installed": installed,
	})
}

func handlePersonaSave(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad body", err)
		return
	}
	if strings.TrimSpace(body.Content) == "" {
		writeError(w, http.StatusBadRequest, "empty content", nil)
		return
	}

	var installed []string
	var warnings []string

	// 1. Claude ~/.claude/CLAUDE.md
	if path, err := writeClaudePersona(body.Content); err != nil {
		warnings = append(warnings, "Claude CLAUDE.md 写入失败："+err.Error())
	} else {
		installed = append(installed, "Claude → "+path)
	}

	// 2. Cursor ~/.cursor/rules/00-persona.mdc
	if path, err := writeCursorPersona(body.Content); err != nil {
		warnings = append(warnings, "Cursor 00-persona.mdc 写入失败："+err.Error())
	} else {
		installed = append(installed, "Cursor → "+path)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"installed": installed,
		"warnings":  warnings,
	})
}

func handlePersonaRemove(w http.ResponseWriter, r *http.Request) {
	var removed []string

	// 1. Claude：只删 marker 之间的段，保留其他内容
	if path, err := removeClaudePersona(); err == nil && path != "" {
		removed = append(removed, path)
	}
	// 2. Cursor：删 .mdc + 删 skills-cursor/persona 整棵树
	home, _ := os.UserHomeDir()
	cursorPath := filepath.Join(home, ".cursor", "rules", "00-persona.mdc")
	if err := os.Remove(cursorPath); err == nil {
		removed = append(removed, cursorPath)
	}
	skillDir := filepath.Join(home, ".cursor", "skills-cursor", "persona")
	if err := os.RemoveAll(skillDir); err == nil {
		if _, se := os.Stat(skillDir); os.IsNotExist(se) {
			removed = append(removed, skillDir)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"removed": removed,
	})
}

// ---------- 底层实现 ----------

func claudeMDPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "CLAUDE.md")
}

// readCurrentPersona 提取 CLAUDE.md 里 marker 之间的内容；没装就返回默认值
func readCurrentPersona() string {
	b, err := os.ReadFile(claudeMDPath())
	if err != nil {
		return personaDefault
	}
	s := string(b)
	i := strings.Index(s, personaMarkerBegin)
	j := strings.Index(s, personaMarkerEnd)
	if i < 0 || j < 0 || j < i {
		return personaDefault
	}
	body := s[i+len(personaMarkerBegin) : j]
	body = strings.TrimSpace(body)
	if body == "" {
		return personaDefault
	}
	return body
}

func claudeMDHasPersona() bool {
	b, err := os.ReadFile(claudeMDPath())
	if err != nil {
		return false
	}
	return strings.Contains(string(b), personaMarkerBegin)
}

func cursorMDCExists() bool {
	home, _ := os.UserHomeDir()
	// 任一位置存在即算"已装"
	if _, err := os.Stat(filepath.Join(home, ".cursor", "rules", "00-persona.mdc")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(home, ".cursor", "skills-cursor", "persona", "SKILL.md")); err == nil {
		return true
	}
	return false
}

// writeClaudePersona 把内容写到 CLAUDE.md 的 marker 段，marker 外的内容保持不变
func writeClaudePersona(content string) (string, error) {
	path := claudeMDPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	block := fmt.Sprintf("\n\n%s\n%s\n%s\n", personaMarkerBegin, strings.TrimSpace(content), personaMarkerEnd)

	existing := ""
	if b, err := os.ReadFile(path); err == nil {
		existing = string(b)
	}

	i := strings.Index(existing, personaMarkerBegin)
	j := strings.Index(existing, personaMarkerEnd)
	var newContent string
	if i >= 0 && j > i {
		// 替换现有段
		before := strings.TrimRight(existing[:i], "\n ")
		after := existing[j+len(personaMarkerEnd):]
		newContent = before + block + strings.TrimLeft(after, "\n ")
	} else {
		// 追加到末尾
		if existing == "" {
			newContent = "# 全局指令（Claude Code User Instructions）\n" + block
		} else {
			newContent = strings.TrimRight(existing, "\n ") + block
		}
	}
	return path, os.WriteFile(path, []byte(newContent), 0644)
}

// removeClaudePersona 删掉 CLAUDE.md 里 marker 之间的段，保留其他
func removeClaudePersona() (string, error) {
	path := claudeMDPath()
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	s := string(b)
	i := strings.Index(s, personaMarkerBegin)
	j := strings.Index(s, personaMarkerEnd)
	if i < 0 || j < i {
		return "", nil // 没装过
	}
	out := strings.TrimRight(s[:i], "\n ") + "\n" + strings.TrimLeft(s[j+len(personaMarkerEnd):], "\n ")
	out = strings.TrimSpace(out)
	if out == "" {
		// 空了就删文件
		_ = os.Remove(path)
		return path, nil
	}
	return path, os.WriteFile(path, []byte(out+"\n"), 0644)
}

func writeCursorPersona(content string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// 1. 老路径 ~/.cursor/rules/00-persona.mdc（rules 格式，有些 Cursor 版本扫）
	rulesDir := filepath.Join(home, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return "", err
	}
	mdcPath := filepath.Join(rulesDir, "00-persona.mdc")
	var mdcSB strings.Builder
	mdcSB.WriteString("---\n")
	mdcSB.WriteString("description: AI 工作态度全局准则 —— 严谨求实 / 不敷衍 / 勤勉 / 交付诚实\n")
	mdcSB.WriteString("alwaysApply: true\n")
	mdcSB.WriteString("---\n\n")
	mdcSB.WriteString(strings.TrimSpace(content))
	mdcSB.WriteString("\n")
	if err := os.WriteFile(mdcPath, []byte(mdcSB.String()), 0644); err != nil {
		return "", err
	}

	// 2. 官方 Cursor Skill 路径 ~/.cursor/skills-cursor/persona/SKILL.md（Cursor 3.x 推荐格式）
	skillDir := filepath.Join(home, ".cursor", "skills-cursor", "persona")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return mdcPath, err
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	var skillSB strings.Builder
	skillSB.WriteString("---\n")
	skillSB.WriteString("name: persona\n")
	skillSB.WriteString("description: AI 工作态度全局准则 —— 严谨求实 / 不敷衍 / 勤勉 / 交付诚实。每次响应前必读。\n")
	skillSB.WriteString("---\n\n")
	skillSB.WriteString(strings.TrimSpace(content))
	skillSB.WriteString("\n")
	if err := os.WriteFile(skillPath, []byte(skillSB.String()), 0644); err != nil {
		return mdcPath, err
	}

	return mdcPath + " + " + skillPath, nil
}
