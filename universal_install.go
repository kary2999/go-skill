package main

// 通用 Agent 适配 —— 让 VSCode 里的各家 AI 插件（Copilot / Cline / Roo /
// Windsurf / Continue / Gemini / DeepSeek-via-Cline 等）都能用上团队规范。
//
// 原理：业界没有统一的 skill 格式，但 AGENTS.md 是事实标准（Codex / Copilot
// agent 模式 / Cline / Roo / Cursor 新版都会读）。其余各家只认自己的 rules
// 文件 —— 我们在项目根目录生成一组"薄适配文件"，内容统一指向 AGENTS.md，
// 规范本体只维护一份。

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// 各家 agent 的 rules 文件位置（相对项目根目录）
var universalAdapters = []struct {
	Agent string // UI 展示名
	Rel   string // 相对路径
}{
	{"GitHub Copilot", ".github/copilot-instructions.md"},
	{"Cline（含 DeepSeek 等模型）", ".clinerules/team-standards.md"},
	{"Roo Code", ".roo/rules/team-standards.md"},
	{"Windsurf", ".windsurfrules"},
	{"Continue", ".continue/rules/team-standards.md"},
	{"Gemini CLI / Code Assist", "GEMINI.md"},
	{"Claude Code", "CLAUDE.md"},
}

const universalPointerContent = `# Team Standards · 团队规范入口

本项目的 AI 协作规范统一维护在根目录的 **AGENTS.md**。

开始任何工作之前，请完整阅读项目根目录的 AGENTS.md 并严格遵循其中的全部约定
（语言、工作流程、汇报格式、暂停机制等）。本文件只是入口指针，规范本体不在这里。
`

// POST /api/universal/install
// body: {"path": "/absolute/path/to/project"}
//
// 1. 项目根目录生成 AGENTS.md（已存在则跳过，与 /api/agents-init 同模板）
// 2. 为每家 agent 生成指向 AGENTS.md 的薄适配文件（已存在则跳过，不覆盖）
func handleUniversalInstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
		writeError(w, http.StatusBadRequest, "path 必填", nil)
		return
	}
	abs, err := filepath.Abs(req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "路径无效", err)
		return
	}
	if st, err := os.Stat(abs); err != nil || !st.IsDir() {
		writeError(w, http.StatusBadRequest, "目录不存在", err)
		return
	}

	var created, skipped, failed []string

	// 1. AGENTS.md 本体
	agentsDst := filepath.Join(abs, "AGENTS.md")
	if _, err := os.Stat(agentsDst); err == nil {
		skipped = append(skipped, "AGENTS.md（已存在，未覆盖）")
	} else if tmpl, err := readEmbedFile("assets/AGENTS.md.template"); err != nil {
		writeError(w, http.StatusInternalServerError, "读 AGENTS.md 模板失败", err)
		return
	} else if err := os.WriteFile(agentsDst, tmpl, 0o644); err != nil {
		failed = append(failed, "AGENTS.md: "+err.Error())
	} else {
		created = append(created, "AGENTS.md")
	}

	// 2. 各家适配文件
	for _, a := range universalAdapters {
		dst := filepath.Join(abs, filepath.FromSlash(a.Rel))
		if _, err := os.Stat(dst); err == nil {
			skipped = append(skipped, a.Rel+"（已存在，未覆盖 · "+a.Agent+"）")
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			failed = append(failed, a.Rel+": "+err.Error())
			continue
		}
		if err := os.WriteFile(dst, []byte(universalPointerContent), 0o644); err != nil {
			failed = append(failed, a.Rel+": "+err.Error())
			continue
		}
		created = append(created, a.Rel+"（"+a.Agent+"）")
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      len(failed) == 0,
		"created": created,
		"skipped": skipped,
		"failed":  failed,
	})
}
